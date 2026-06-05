package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/artem/url-shortener/internal/geo"
	"github.com/artem/url-shortener/internal/model"
	"github.com/artem/url-shortener/internal/service"
)

type mockLinkRepo struct{}

func (m *mockLinkRepo) Insert(url string, ownerID *string) (*model.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) UpdateShortCode(id int64, code string) error { return nil }
func (m *mockLinkRepo) FindByShortCode(code string) (*model.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) FindByID(id int64) (*model.Link, error) { return nil, nil }
func (m *mockLinkRepo) FindByOwner(ownerID string) ([]model.Link, error) {
	return nil, nil
}
func (m *mockLinkRepo) Delete(id int64) error                     { return nil }
func (m *mockLinkRepo) InsertClick(click *model.Click) error       { return nil }
func (m *mockLinkRepo) GetStats(linkID int64) (*model.LinkStats, error) {
	return nil, nil
}

type mockCacheRepo struct{}

func (m *mockCacheRepo) Get(code string) (string, bool) { return "", false }
func (m *mockCacheRepo) Set(code, originalURL string, ttl int) error { return nil }
func (m *mockCacheRepo) DeleteCache(code string) error { return nil }

func TestRedirectHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cache := &mockCacheRepo{}
	geoClient := geo.NewMockClient("", "")
	svc := service.NewShortenerService(&mockLinkRepo{}, cache)
	h := NewRedirectHandler(svc, &mockLinkRepo{}, geoClient)
	r.GET("/:code", h.Redirect)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRedirectHandler_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cache := &mockCacheRepo{}
	geoClient := geo.NewMockClient("", "")
	svc := service.NewShortenerService(&mockLinkRepo{}, cache)
	h := NewRedirectHandler(svc, &mockLinkRepo{}, geoClient)
	r.GET("/", h.Redirect)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
