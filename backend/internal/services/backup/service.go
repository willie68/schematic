package backup

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"github.com/willie68/gowillie68/pkg/fileutils"
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

type service struct {
	duration time.Duration
	path     string
	log      *slog.Logger
	store    store
	blob     blobStore
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func New(inj do.Injector, option ...func(*service)) (*service, error) {
	s := &service{
		log:   logging.New("backup"),
		store: do.MustInvokeAs[store](inj),
		blob:  do.MustInvokeAs[blobStore](inj),
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
	if s.path == "" {
		return errors.New("backup path is empty")
	}
	if s.store == nil {
		return errors.New("backup store is not initialised")
	}
	if s.blob == nil {
		return errors.New("backup blob store is not initialised")
	}

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

	s.log.Info("backup completed", "path", backupDir)
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

// Shutdown stops the backup service gracefully. Needed to implement the do.Shutdownable interface.
func (s *service) Shutdown() {
	_ = s.Stop()
}
