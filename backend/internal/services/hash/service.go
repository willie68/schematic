package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"

	"github.com/samber/do/v2"
	"github.com/willie68/schematics2/backend/internal/domain/model"
	"github.com/willie68/schematics2/backend/internal/logging"
)

type storage interface {
	All(f func(m model.Document) bool) error
	Upsert(doc model.Document) error
}

type blobStore interface {
	Load(container *model.ContainerInfo) ([]byte, error)
}

type service struct {
	stg  storage
	blob blobStore
	log  *slog.Logger
}

func New(inj do.Injector) *service {
	return &service{
		stg:  do.MustInvokeAs[storage](inj),
		blob: do.MustInvokeAs[blobStore](inj),
		log:  logging.New("rebuild_hashes"),
	}
}

// RebuildAllHashes iterates over all documents and rebuilds the hashes for files that have an empty hash.
func (s *service) RebuildAllHashes() error {
	err := s.stg.All(func(doc model.Document) bool {
		rebuilded := false
		for x, file := range doc.Files {
			if file.Hash == "" {
				s.log.Info("rebuild hash for document", "docid", doc.ID, "file", file.Name)

				hash, err := s.GetHash(file)
				if err != nil {
					s.log.Error("failed to rebuild hash for document", "docid", doc.ID, "file", file.Name, "error", err)
				}
				if hash == "" {
					s.log.Error("failed to rebuild hash for document", "docid", doc.ID, "file", file.Name, "error", "hash is empty")
					continue
				}
				file.Hash = hash
				doc.Files[x] = file
				rebuilded = true
			}
		}
		if rebuilded {
			err := s.stg.Upsert(doc)
			if err != nil {
				s.log.Error("failed to update document", "docid", doc.ID, "error", err)
			}
		}
		return true
	})
	return err
}

func (s *service) GetHash(file model.DocumentFile) (string, error) {
	payload, err := s.blob.Load(file.Container)
	if err != nil {
		if file.Container == nil {
			s.log.Error("skip document file: no container info", "file:", file.Name)
			return "", err
		}
		s.log.Error("skip document file: load failed", "container", file.Container.ContainerNumber, "index", file.Container.Offset, "error", err)
		return "", err
	}
	hash := sha256.Sum256(payload)
	return "sha2:" + hex.EncodeToString(hash[:]), nil
}

func (s *service) GetHashFromPayload(payload []byte) string {
	hash := sha256.Sum256(payload)
	return "sha2:" + hex.EncodeToString(hash[:])
}

func (s *service) GetHashFromFile(filePath string) string {
	payload, err := os.ReadFile(filePath)
	if err != nil {
		s.log.Error("failed to read file", "file", filePath, "error", err)
		return ""
	}
	return s.GetHashFromPayload(payload)
}
