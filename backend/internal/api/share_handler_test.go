package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/willie68/schematics2/backend/internal/domain/model"
)

type stubShareService struct {
	createFn func(ctx context.Context, share *model.Share) (string, error)
	getFn    func(ctx context.Context, link string) (*model.Share, error)
}

func (s stubShareService) CreateShare(ctx context.Context, share *model.Share) (string, error) {
	if s.createFn == nil {
		return "", errors.New("create not implemented")
	}
	return s.createFn(ctx, share)
}

func (s stubShareService) GetShare(ctx context.Context, link string) (*model.Share, error) {
	if s.getFn == nil {
		return nil, errors.New("get not implemented")
	}
	return s.getFn(ctx, link)
}

func withRouteParam(req *http.Request, key, value string) *http.Request {
	rctx, _ := req.Context().Value(chi.RouteCtxKey).(*chi.Context)
	if rctx == nil {
		rctx = chi.NewRouteContext()
	}
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestGetSharedDocument_Success(t *testing.T) {
	mockDocStore := newMockdocumentStore(t)
	mockDocStore.EXPECT().GetByID(mock.Anything, "doc-1").Return(model.Document{ID: "doc-1", PrivateFile: true, Owner: "owner"}, nil)

	h := &Handler{
		docStore: mockDocStore,
		shareSvc: stubShareService{
			getFn: func(_ context.Context, link string) (*model.Share, error) {
				assert.Equal(t, "share-1", link)
				return &model.Share{Link: "share-1", DocumentID: "doc-1", ValidTo: time.Now().Add(time.Hour)}, nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/shares/share-1", nil)
	req = withRouteParam(req, "link", "share-1")
	w := httptest.NewRecorder()

	h.getSharedDocument(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, "doc-1", body["id"])
}

func TestGetSharedDocument_Expired(t *testing.T) {
	h := &Handler{
		shareSvc: stubShareService{
			getFn: func(_ context.Context, _ string) (*model.Share, error) {
				return &model.Share{Link: "share-1", DocumentID: "doc-1", ValidTo: time.Now().Add(-time.Minute)}, nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/shares/share-1", nil)
	req = withRouteParam(req, "link", "share-1")
	w := httptest.NewRecorder()

	h.getSharedDocument(w, req)

	require.Equal(t, http.StatusGone, w.Code)
	assert.Contains(t, w.Body.String(), "share expired")
}

func TestDownloadSharedFile_Success(t *testing.T) {
	mockDocStore := newMockdocumentStore(t)
	mockBlobStore := newMockblobStore(t)

	doc := model.Document{
		ID: "doc-1",
		Files: []model.DocumentFile{{
			Name:      "manual.pdf",
			Type:      "manual",
			MIMEType:  "application/pdf",
			Page:      1,
			Container: &model.ContainerInfo{ID: "cont-1"},
		}},
	}

	mockDocStore.EXPECT().GetByID(mock.Anything, "doc-1").Return(doc, nil)
	mockBlobStore.EXPECT().Load(&model.ContainerInfo{ID: "cont-1"}).Return([]byte("file-bytes"), nil)

	h := &Handler{
		docStore: mockDocStore,
		blob:     mockBlobStore,
		log:      testLogger(),
		shareSvc: stubShareService{
			getFn: func(_ context.Context, _ string) (*model.Share, error) {
				return &model.Share{Link: "share-1", DocumentID: "doc-1", ValidTo: time.Now().Add(time.Hour)}, nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/shares/share-1/files/manual.pdf", nil)
	req = withRouteParam(req, "link", "share-1")
	req = withRouteParam(req, "filename", "manual.pdf")
	w := httptest.NewRecorder()

	h.downloadSharedFile(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, "manual.pdf", body["name"])
	assert.Equal(t, "application/pdf", body["mimetype"])
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("file-bytes")), body["data"])
}

func TestCreateShare_DocumentNotFound(t *testing.T) {
	mockDocStore := newMockdocumentStore(t)
	mockDocStore.EXPECT().HasID(mock.Anything, "doc-1").Return(false)

	shareCreated := false
	h := &Handler{
		docStore: mockDocStore,
		shareSvc: stubShareService{
			createFn: func(_ context.Context, _ *model.Share) (string, error) {
				shareCreated = true
				return "", nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/shares", nil)
	req = withRouteParam(req, "id", "doc-1")
	ctx := context.WithValue(req.Context(), ctxSubjectKey{}, "alice")
	ctx = context.WithValue(ctx, ctxRolesKey{}, []string{"user"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.createShare(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	assert.False(t, shareCreated)
}

func TestCreateShare_Success(t *testing.T) {
	mockDocStore := newMockdocumentStore(t)
	mockDocStore.EXPECT().HasID(mock.Anything, "doc-1").Return(true)

	h := &Handler{
		docStore: mockDocStore,
		shareSvc: stubShareService{
			createFn: func(_ context.Context, share *model.Share) (string, error) {
				require.Equal(t, "doc-1", share.DocumentID)
				require.Equal(t, "alice", share.Owner)
				require.True(t, share.ValidTo.After(time.Now()))
				return "685e97d7b3f1c8aa4f4a1234", nil
			},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/doc-1/shares", nil)
	req.Host = "example.org"
	req = withRouteParam(req, "id", "doc-1")
	ctx := context.WithValue(req.Context(), ctxSubjectKey{}, "alice")
	ctx = context.WithValue(ctx, ctxRolesKey{}, []string{"user"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.createShare(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, "created", body["status"])
	assert.NotEmpty(t, body["id"])
	link, ok := body["link"].(string)
	require.True(t, ok)
	assert.Contains(t, link, "example.org")
	assert.NotContains(t, link, "id=")
	assert.Contains(t, link, "share=685e97d7b3f1c8aa4f4a1234")
	assert.True(t, strings.Contains(link, "/client/search?"))
}
