package hash

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willie68/schematics2/backend/internal/domain/model"
)

type fakeStorage struct {
	documents []model.Document
	upserts   []model.Document
	upsertErr error
	allErr    error
	allCalls  int
}

func (f *fakeStorage) All(fn func(m model.Document) bool) error {
	f.allCalls++
	if f.allErr != nil {
		return f.allErr
	}
	for _, doc := range f.documents {
		if !fn(doc) {
			break
		}
	}
	return nil
}

func (f *fakeStorage) Upsert(doc model.Document) error {
	f.upserts = append(f.upserts, doc)
	return f.upsertErr
}

type fakeBlobStore struct {
	payload []byte
	err     error
	loaded  []*model.ContainerInfo
}

func (f *fakeBlobStore) Load(container *model.ContainerInfo) ([]byte, error) {
	f.loaded = append(f.loaded, container)
	if f.err != nil {
		return nil, f.err
	}
	return f.payload, nil
}

func TestGetHashFromPayload(t *testing.T) {
	t.Parallel()

	svc := &service{log: slog.Default()}

	hash := svc.GetHashFromPayload([]byte("hello world"))

	assert.Equal(t, "sha2:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", hash)
}

func TestGetHash(t *testing.T) {
	t.Parallel()

	blob := &fakeBlobStore{payload: []byte("hello world")}
	container := &model.ContainerInfo{ContainerNumber: 12, Offset: 99}
	svc := &service{blob: blob, log: slog.Default()}

	hash, err := svc.GetHash(model.DocumentFile{Container: container})

	require.NoError(t, err)
	assert.Equal(t, "sha2:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", hash)
	require.Len(t, blob.loaded, 1)
	assert.Same(t, container, blob.loaded[0])
}

func TestGetHashReturnsErrorWhenBlobLoadFails(t *testing.T) {
	t.Parallel()

	loadErr := errors.New("blob load failed")
	blob := &fakeBlobStore{err: loadErr}
	svc := &service{blob: blob, log: slog.Default()}

	hash, err := svc.GetHash(model.DocumentFile{Container: &model.ContainerInfo{}})

	require.ErrorIs(t, err, loadErr)
	assert.Empty(t, hash)
}

func TestGetHashFromFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "payload.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello world"), 0o600))

	svc := &service{log: slog.Default()}

	hash := svc.GetHashFromFile(filePath)

	assert.Equal(t, "sha2:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", hash)
}

func TestGetHashFromFileReturnsEmptyStringOnReadError(t *testing.T) {
	t.Parallel()

	svc := &service{log: slog.Default()}

	hash := svc.GetHashFromFile(filepath.Join(t.TempDir(), "missing.txt"))

	assert.Empty(t, hash)
}

func TestRebuildAllHashesUpdatesOnlyMissingHashes(t *testing.T) {
	t.Parallel()

	stg := &fakeStorage{
		documents: []model.Document{
			{
				ID: "doc-1",
				Files: []model.DocumentFile{
					{
						Name:      "already-hashed.pdf",
						Hash:      "sha2:existing",
						Container: &model.ContainerInfo{ContainerNumber: 1, Offset: 10},
					},
					{
						Name:      "missing-hash.pdf",
						Container: &model.ContainerInfo{ContainerNumber: 2, Offset: 20},
					},
				},
			},
			{
				ID: "doc-2",
				Files: []model.DocumentFile{
					{
						Name:      "missing-hash-2.pdf",
						Container: &model.ContainerInfo{ContainerNumber: 3, Offset: 30},
					},
				},
			},
		},
	}
	blob := &fakeBlobStore{payload: []byte("hello world")}
	svc := &service{stg: stg, blob: blob, log: slog.Default()}

	err := svc.RebuildAllHashes()

	require.NoError(t, err)
	assert.Equal(t, 1, stg.allCalls)
	require.Len(t, stg.upserts, 2)
	assert.Equal(t, "sha2:existing", stg.upserts[0].Files[0].Hash)
	assert.Equal(t, "sha2:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", stg.upserts[0].Files[1].Hash)
	assert.Equal(t, "sha2:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", stg.upserts[1].Files[0].Hash)
	require.Len(t, blob.loaded, 2)
	assert.Equal(t, 2, blob.loaded[0].ContainerNumber)
	assert.Equal(t, 3, blob.loaded[1].ContainerNumber)
}

func TestRebuildAllHashesSkipsUpsertWhenHashRebuildFails(t *testing.T) {
	t.Parallel()

	stg := &fakeStorage{
		documents: []model.Document{
			{
				ID: "doc-1",
				Files: []model.DocumentFile{{
					Name:      "broken.pdf",
					Container: &model.ContainerInfo{ContainerNumber: 4, Offset: 40},
				}},
			},
		},
	}
	blob := &fakeBlobStore{err: errors.New("blob load failed")}
	svc := &service{stg: stg, blob: blob, log: slog.Default()}

	err := svc.RebuildAllHashes()

	require.NoError(t, err)
	assert.Empty(t, stg.upserts)
}

func TestRebuildAllHashesReturnsStorageAllError(t *testing.T) {
	t.Parallel()

	allErr := errors.New("storage all failed")
	stg := &fakeStorage{allErr: allErr}
	svc := &service{stg: stg, blob: &fakeBlobStore{}, log: slog.Default()}

	err := svc.RebuildAllHashes()

	require.ErrorIs(t, err, allErr)
	assert.Empty(t, stg.upserts)
}
