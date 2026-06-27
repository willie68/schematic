package backup

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"github.com/willie68/gowillie68/pkg/fileutils"
	"github.com/willie68/gowillie68/pkg/measurement"
	"github.com/willie68/schematics2/backend/internal/domain/model"
	"github.com/willie68/schematics2/backend/internal/logging"
)

type store interface {
	ExportBackup(ctx context.Context, destPath string) error
}

type blobStore interface {
	Load(ci *model.ContainerInfo) ([]byte, error)
}

type Option struct {
}
type measurementService interface {
	Start(name string) measurement.Monitor
}

const maxZipPartSize int64 = 500 * 1024 * 1024

type service struct {
	duration    time.Duration
	path        string
	log         *slog.Logger
	store       store
	blob        blobStore
	measurement measurementService
	backup      bool
	stopCh      chan struct{}
	doneCh      chan struct{}
}

func New(inj do.Injector, option ...func(*service)) (*service, error) {
	s := &service{
		log:         logging.New("backup"),
		store:       do.MustInvokeAs[store](inj),
		blob:        do.MustInvokeAs[blobStore](inj),
		measurement: do.MustInvokeAs[measurementService](inj),
	}
	for _, opt := range option {
		opt(s)
	}
	return s, nil
}

func WithPath(backuppath string) func(*service) {
	return func(s *service) {
		s.log.Info("set backup path", "path", backuppath)
		s.path = backuppath
	}
}

func WithDuration(duration time.Duration) func(*service) {
	return func(s *service) {
		s.log.Info("set backup duration", "duration", duration)
		s.duration = duration
	}
}

func WithBackup(backup bool) func(*service) {
	return func(s *service) {
		s.log.Info("set backup", "backup", backup)
		s.backup = backup
	}
}

func (s *service) Start() error {
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})

	go func() {
		defer close(s.doneCh)
		firstBackupTimer := time.NewTimer(10 * time.Second)
		defer firstBackupTimer.Stop()

		select {
		case <-s.stopCh:
			return
		case <-firstBackupTimer.C:
			if err := s.Backup(); err != nil {
				s.log.Error("backup failed", "error", err)
			}
		}

		ticker := time.NewTicker(s.duration)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				if err := s.Backup(); err != nil {
					s.log.Error("backup failed", "error", err)
				}
			}
		}
	}()

	return nil
}

func (s *service) Stop() error {
	if s.stopCh != nil {
		close(s.stopCh)
		<-s.doneCh
	}
	return nil
}

func (s *service) Backup() error {
	if !s.backup {
		s.log.Info("backup is disabled, skipping")
		return nil
	}
	if s.path == "" {
		return errors.New("backup path is empty")
	}
	if s.store == nil {
		return errors.New("backup store is not initialised")
	}
	if s.blob == nil {
		return errors.New("backup blob store is not initialised")
	}

	mon := s.measurement.Start("backup")
	defer mon.Stop()

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	backupDir := filepath.Join(s.path, timestamp)
	dbDir := filepath.Join(backupDir, "db")

	if err := s.store.ExportBackup(context.Background(), dbDir); err != nil {
		return fmt.Errorf("export mongo collections: %w", err)
	}
	if err := s.exportEffectImages(dbDir); err != nil {
		return fmt.Errorf("export effect images: %w", err)
	}
	if err := s.exportDocumentFiles(dbDir); err != nil {
		return fmt.Errorf("export document files: %w", err)
	}
	archivePaths, err := archiveBackupDirectory(backupDir, maxZipPartSize)
	if err != nil {
		return fmt.Errorf("archive backup directory: %w", err)
	}
	if err := os.RemoveAll(backupDir); err != nil {
		return fmt.Errorf("remove backup directory after archiving: %w", err)
	}
	if err := cleanupOldBackupArchives(s.path, archivePaths); err != nil {
		return fmt.Errorf("cleanup old backup archives: %w", err)
	}

	s.log.Info("backup completed", "archives", archivePaths)
	return nil
}

// exportEffectImages reads the effects collection from the exported backup, extracts image references and saves the corresponding blobs to the backup directory.
func (s *service) exportEffectImages(dbDir string) error {
	effectsFile, err := os.Open(filepath.Join(dbDir, "effects.json"))
	if err != nil {
		return err
	}
	defer effectsFile.Close()

	imagesDir := filepath.Join(dbDir, "images")
	if err = os.MkdirAll(imagesDir, 0o755); err != nil {
		return err
	}

	decoder := json.NewDecoder(bufio.NewReader(effectsFile))
	if err = validateJSONArrayStart(decoder, "effects"); err != nil {
		return err
	}

	type effectExport struct {
		ID    string               `json:"_id"`
		Image *model.ContainerInfo `json:"image"`
	}

	for decoder.More() {
		var effect effectExport
		if err = decoder.Decode(&effect); err != nil {
			return fmt.Errorf("decode effect entry: %w", err)
		}
		if err = s.exportSingleEffectImage(effect.ID, effect.Image, imagesDir); err != nil {
			return err
		}
	}

	if err = validateJSONArrayEnd(decoder, "effects"); err != nil {
		return err
	}

	return nil
}

// exportDocumentFiles streams documents.json, extracts all file references and
// writes the corresponding blobs to:
//
//	db/documents/<manufacturer>/<model>/<index+1 zero-padded>_<filename>
func (s *service) exportDocumentFiles(dbDir string) error {
	docsFile, err := os.Open(filepath.Join(dbDir, "documents.json"))
	if err != nil {
		return err
	}
	defer docsFile.Close()

	decoder := json.NewDecoder(bufio.NewReader(docsFile))
	if err = validateJSONArrayStart(decoder, "documents"); err != nil {
		return err
	}

	type docExport struct {
		ID           string               `json:"_id"`
		Manufacturer string               `json:"manufacturer"`
		Model        string               `json:"model"`
		Files        []model.DocumentFile `json:"files"`
	}

	for decoder.More() {
		var doc docExport
		if err = decoder.Decode(&doc); err != nil {
			s.log.Error("skip document: decode failed", "error", err)
			continue
		}
		s.exportSingleDocumentFiles(dbDir, doc.ID, doc.Manufacturer, doc.Model, doc.Files)
	}

	if err = validateJSONArrayEnd(decoder, "documents"); err != nil {
		return err
	}

	return nil
}

func validateJSONArrayStart(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("read %s array token: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '[' {
		return fmt.Errorf("%s.json has invalid root JSON value", name)
	}
	return nil
}

func validateJSONArrayEnd(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("%s.json ended before closing array", name)
		}
		return fmt.Errorf("read closing %s array token: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != ']' {
		return fmt.Errorf("%s.json has invalid closing JSON token", name)
	}
	return nil
}

func (s *service) exportSingleEffectImage(effectID string, image *model.ContainerInfo, imagesDir string) error {
	if image == nil {
		return nil
	}

	name := filepath.Base(strings.TrimSpace(image.Name))
	if name == "" || name == "." {
		s.log.Warn("skip effect image without filename", "effectId", effectID)
		return nil
	}

	payload, err := s.blob.Load(image)
	if err != nil {
		return fmt.Errorf("load image for effect %q: %w", effectID, err)
	}

	if err = os.WriteFile(filepath.Join(imagesDir, name), payload, 0o644); err != nil {
		return fmt.Errorf("write image %q for effect %q: %w", name, effectID, err)
	}

	return nil
}

func (s *service) exportSingleDocumentFiles(dbDir, docID, manufacturer, modelName string, files []model.DocumentFile) {
	if len(files) == 0 {
		return
	}

	destDir, err := prepareDocumentDir(dbDir, manufacturer, modelName)
	if err != nil {
		s.log.Error("skip document: create dir failed", "docId", docID, "error", err)
		return
	}

	for i, f := range files {
		s.exportSingleDocumentFile(docID, i, destDir, f)
	}
}

func prepareDocumentDir(dbDir, manufacturer, modelName string) (string, error) {
	manufacturerDir := fileutils.SanitizePathSegment(manufacturer)
	modelDir := fileutils.SanitizePathSegment(modelName)
	destDir := filepath.Join(dbDir, "documents", manufacturerDir, modelDir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("create documents dir for %q/%q: %w", manufacturer, modelName, err)
	}
	return destDir, nil
}

func (s *service) exportSingleDocumentFile(docID string, index int, destDir string, file model.DocumentFile) {
	if file.Container == nil {
		return
	}

	payload, err := s.blob.Load(file.Container)
	if err != nil {
		s.log.Error("skip document file: load failed", "docId", docID, "index", index, "error", err)
		return
	}

	baseName := fileutils.SanitizeFileName(file.Name)
	if baseName == "" {
		s.log.Warn("skip document file without filename", "docId", docID, "index", index)
		return
	}

	fileName := fmt.Sprintf("%02d_%s", index+1, baseName)
	if err = os.WriteFile(filepath.Join(destDir, fileName), payload, 0o644); err != nil {
		s.log.Error("skip document file: write failed", "docId", docID, "file", fileName, "error", err)
	}
}

func archiveBackupDirectory(dir string, maxPartSize int64) (archivePaths []string, err error) {
	files, err := collectBackupFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("backup directory %q is empty", dir)
	}

	defer func() {
		if err == nil {
			return
		}
		for _, archivePath := range archivePaths {
			_ = os.Remove(archivePath)
		}
	}()

	baseName := filepath.Base(dir)
	baseDir := filepath.Dir(dir)
	partIndex := 1
	archive, err := newZipPartWriter(baseDir, baseName, partIndex)
	if err != nil {
		return nil, err
	}
	archivePaths = append(archivePaths, archive.path)

	for _, file := range files {
		if archive.shouldRotate(file.size, maxPartSize) {
			if err = archive.close(); err != nil {
				return archivePaths, err
			}
			partIndex++
			archive, err = newZipPartWriter(baseDir, baseName, partIndex)
			if err != nil {
				return archivePaths, err
			}
			archivePaths = append(archivePaths, archive.path)
		}

		if err = archive.addFile(file); err != nil {
			_ = archive.close()
			return archivePaths, err
		}
	}

	if err = archive.close(); err != nil {
		return archivePaths, err
	}

	return archivePaths, nil
}

func cleanupOldBackupArchives(backupRoot string, keepArchives []string) error {
	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return fmt.Errorf("read backup root: %w", err)
	}

	keepSet := make(map[string]struct{}, len(keepArchives))
	for _, archive := range keepArchives {
		keepSet[filepath.Clean(archive)] = struct{}{}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isBackupArchiveFile(name) {
			continue
		}

		archivePath := filepath.Clean(filepath.Join(backupRoot, name))
		if _, keep := keepSet[archivePath]; keep {
			continue
		}

		if err := os.Remove(archivePath); err != nil {
			return fmt.Errorf("remove old archive %q: %w", archivePath, err)
		}
	}

	return nil
}

func isBackupArchiveFile(name string) bool {
	if !strings.HasSuffix(strings.ToLower(name), ".zip") {
		return false
	}

	base := strings.TrimSuffix(name, ".zip")
	parts := strings.Split(base, ".part")
	if len(parts) != 2 {
		return false
	}

	ts := parts[0]
	idx := parts[1]
	if len(ts) != 16 || len(idx) != 3 {
		return false
	}
	if ts[8] != 'T' || ts[15] != 'Z' {
		return false
	}

	for i := 0; i < len(ts); i++ {
		if i == 8 || i == 15 {
			continue
		}
		if ts[i] < '0' || ts[i] > '9' {
			return false
		}
	}
	for i := 0; i < len(idx); i++ {
		if idx[i] < '0' || idx[i] > '9' {
			return false
		}
	}

	return true
}

type backupFile struct {
	absPath string
	relPath string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

type zipPartWriter struct {
	path         string
	file         *os.File
	zipWriter    *zip.Writer
	currentSize  int64
	entriesCount int
	closed       bool
}

func collectBackupFiles(dir string) ([]backupFile, error) {
	files := make([]backupFile, 0)
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, backupFile{
			absPath: path,
			relPath: filepath.ToSlash(relPath),
			size:    info.Size(),
			mode:    info.Mode(),
			modTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect backup files: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].relPath < files[j].relPath
	})

	return files, nil
}

func newZipPartWriter(baseDir, baseName string, partIndex int) (*zipPartWriter, error) {
	archivePath := filepath.Join(baseDir, fmt.Sprintf("%s.part%03d.zip", baseName, partIndex))
	file, err := os.Create(archivePath)
	if err != nil {
		return nil, fmt.Errorf("create archive part %q: %w", archivePath, err)
	}

	return &zipPartWriter{
		path:      archivePath,
		file:      file,
		zipWriter: zip.NewWriter(file),
	}, nil
}

func (z *zipPartWriter) shouldRotate(nextFileSize, maxPartSize int64) bool {
	if z.entriesCount == 0 {
		return false
	}
	estimatedNextSize := z.currentSize + nextFileSize + 16*1024
	return estimatedNextSize > maxPartSize
}

func (z *zipPartWriter) addFile(file backupFile) error {
	source, err := os.Open(file.absPath)
	if err != nil {
		return fmt.Errorf("open backup file %q: %w", file.relPath, err)
	}
	defer source.Close()

	header, err := zip.FileInfoHeader(fileInfoAdapter{file: file})
	if err != nil {
		return fmt.Errorf("build zip header for %q: %w", file.relPath, err)
	}
	header.Name = file.relPath
	header.Method = zip.Deflate

	writer, err := z.zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("create zip entry for %q: %w", file.relPath, err)
	}

	written, err := io.Copy(writer, source)
	if err != nil {
		return fmt.Errorf("write zip entry for %q: %w", file.relPath, err)
	}

	z.entriesCount++
	z.currentSize += written + 16*1024
	return nil
}

func (z *zipPartWriter) close() error {
	if z.closed {
		return nil
	}
	z.closed = true

	if err := z.zipWriter.Close(); err != nil {
		_ = z.file.Close()
		return fmt.Errorf("close zip writer for %q: %w", z.path, err)
	}
	if err := z.file.Close(); err != nil {
		return fmt.Errorf("close archive file %q: %w", z.path, err)
	}
	return nil
}

type fileInfoAdapter struct {
	file backupFile
}

func (f fileInfoAdapter) Name() string       { return filepath.Base(f.file.relPath) }
func (f fileInfoAdapter) Size() int64        { return f.file.size }
func (f fileInfoAdapter) Mode() os.FileMode  { return f.file.mode }
func (f fileInfoAdapter) ModTime() time.Time { return f.file.modTime }
func (f fileInfoAdapter) IsDir() bool        { return false }
func (f fileInfoAdapter) Sys() any           { return nil }

// Shutdown stops the backup service gracefully. Needed to implement the do.Shutdownable interface.
func (s *service) Shutdown() {
	_ = s.Stop()
}
