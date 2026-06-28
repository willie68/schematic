package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/willie68/schematics2/backend/internal/domain/model"
)

type presenceHasherStub struct{}

func (presenceHasherStub) GetHashFromPayload(payload []byte) string {
	if string(payload) == "hello" {
		return "sha2:test"
	}
	return "sha2:other"
}

type presenceDocStoreStub struct {
	hasHash bool
}

func (s presenceDocStoreStub) Upsert(doc model.Document) error                   { return nil }
func (s presenceDocStoreStub) ListTags(ctx context.Context) ([]model.Tag, error) { return nil, nil }
func (s presenceDocStoreStub) SuggestTags(ctx context.Context, prefix string, limit int) ([]model.Tag, error) {
	return nil, nil
}
func (s presenceDocStoreStub) SuggestManufacturers(ctx context.Context, prefix string, limit int) ([]string, error) {
	return nil, nil
}
func (s presenceDocStoreStub) HasID(ctx context.Context, id string) bool { return false }
func (s presenceDocStoreStub) GetByID(ctx context.Context, id string) (model.Document, error) {
	return model.Document{}, nil
}
func (s presenceDocStoreStub) DeleteByID(ctx context.Context, id string) error { return nil }
func (s presenceDocStoreStub) CountAll(ctx context.Context) (int64, error)     { return 0, nil }
func (s presenceDocStoreStub) HasHash(ctx context.Context, hash string) bool {
	if hash != "sha2:test" {
		return false
	}
	return s.hasHash
}

func TestCheckFilePresence_Found(t *testing.T) {
	h := &Handler{
		docStore: presenceDocStoreStub{hasHash: true},
		hasher:   presenceHasherStub{},
	}

	payload := base64.StdEncoding.EncodeToString([]byte("hello"))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/presence", strings.NewReader(`{"data":"`+payload+`"}`))
	w := httptest.NewRecorder()

	h.checkFilePresence(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, true, body["presence"])
}

func TestCheckFilePresence_NotFound(t *testing.T) {
	h := &Handler{
		docStore: presenceDocStoreStub{hasHash: false},
		hasher:   presenceHasherStub{},
	}

	payload := base64.StdEncoding.EncodeToString([]byte("hello"))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/presence", strings.NewReader(`{"payload":"`+payload+`"}`))
	w := httptest.NewRecorder()

	h.checkFilePresence(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, false, body["presence"])
}

func TestCheckFilePresence_InvalidPayload(t *testing.T) {
	h := &Handler{
		docStore: presenceDocStoreStub{hasHash: false},
		hasher:   presenceHasherStub{},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/presence", strings.NewReader(`{"data":"%%%"}`))
	w := httptest.NewRecorder()

	h.checkFilePresence(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid payload")
}
