package service

import (
	"errors"
	"testing"

	"github.com/artem/url-shortener/internal/model"
)

type mockLinkRepo struct {
	insertFunc  func(url string, ownerID *string) (*model.Link, error)
	updateFunc  func(id int64, code string) error
	findByCode  func(code string) (*model.Link, error)
	findByID    func(id int64) (*model.Link, error)
	findByOwner func(ownerID string) ([]model.Link, error)
	deleteFunc  func(id int64) error
	clickFunc   func(click *model.Click) error
	statsFunc   func(linkID int64) (*model.LinkStats, error)
}

func (m *mockLinkRepo) Insert(url string, ownerID *string) (*model.Link, error) {
	return m.insertFunc(url, ownerID)
}
func (m *mockLinkRepo) UpdateShortCode(id int64, code string) error {
	return m.updateFunc(id, code)
}
func (m *mockLinkRepo) FindByShortCode(code string) (*model.Link, error) {
	return m.findByCode(code)
}
func (m *mockLinkRepo) FindByID(id int64) (*model.Link, error) {
	return m.findByID(id)
}
func (m *mockLinkRepo) FindByOwner(ownerID string) ([]model.Link, error) {
	return m.findByOwner(ownerID)
}
func (m *mockLinkRepo) Delete(id int64) error { return m.deleteFunc(id) }
func (m *mockLinkRepo) InsertClick(click *model.Click) error {
	return m.clickFunc(click)
}
func (m *mockLinkRepo) GetStats(linkID int64) (*model.LinkStats, error) {
	return m.statsFunc(linkID)
}

type mockCacheRepo struct {
	getFunc func(code string) (string, bool)
	setFunc func(code, originalURL string, ttl int) error
	delFunc func(code string) error
}

func (m *mockCacheRepo) Get(code string) (string, bool) {
	return m.getFunc(code)
}
func (m *mockCacheRepo) Set(code, originalURL string, ttl int) error {
	return m.setFunc(code, originalURL, ttl)
}
func (m *mockCacheRepo) DeleteCache(code string) error { return m.delFunc(code) }

func TestCreate_InvalidURL(t *testing.T) {
	svc := NewShortenerService(nil, nil)
	_, err := svc.Create("not-a-url", nil, "http://localhost:8080")
	if !errors.Is(err, ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestCreate_EmptyURL(t *testing.T) {
	svc := NewShortenerService(nil, nil)
	_, err := svc.Create("", nil, "http://localhost:8080")
	if !errors.Is(err, ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestCreate_OnlySpaces(t *testing.T) {
	svc := NewShortenerService(nil, nil)
	_, err := svc.Create("   ", nil, "http://localhost:8080")
	if !errors.Is(err, ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestCreate_MissingScheme(t *testing.T) {
	svc := NewShortenerService(nil, nil)
	_, err := svc.Create("example.com/test", nil, "http://localhost:8080")
	if !errors.Is(err, ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestCreate_FTPNotAllowed(t *testing.T) {
	svc := NewShortenerService(nil, nil)
	_, err := svc.Create("ftp://example.com", nil, "http://localhost:8080")
	if !errors.Is(err, ErrInvalidURL) {
		t.Errorf("expected ErrInvalidURL, got %v", err)
	}
}

func TestCreate_ValidURL(t *testing.T) {
	linkRepo := &mockLinkRepo{
		insertFunc: func(url string, ownerID *string) (*model.Link, error) {
			return &model.Link{ID: 42}, nil
		},
		updateFunc: func(id int64, code string) error {
			if code != "g" {
				t.Errorf("expected code 'g', got %q", code)
			}
			return nil
		},
	}
	svc := NewShortenerService(linkRepo, nil)
	resp, err := svc.Create("https://example.com", nil, "http://localhost:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ShortCode != "g" {
		t.Errorf("got short_code %q, want %q", resp.ShortCode, "g")
	}
	if resp.OriginalURL != "https://example.com" {
		t.Errorf("got original_url %q, want %q", resp.OriginalURL, "https://example.com")
	}
	if resp.ShortURL != "http://localhost:8080/g" {
		t.Errorf("got short_url %q, want %q", resp.ShortURL, "http://localhost:8080/g")
	}
}

func TestCreate_WithOwnerID(t *testing.T) {
	owner := "user-42"
	linkRepo := &mockLinkRepo{
		insertFunc: func(url string, ownerID *string) (*model.Link, error) {
			if ownerID == nil || *ownerID != owner {
				t.Errorf("expected owner %q, got %v", owner, ownerID)
			}
			return &model.Link{ID: 100}, nil
		},
		updateFunc: func(id int64, code string) error { return nil },
	}
	svc := NewShortenerService(linkRepo, nil)
	resp, err := svc.Create("https://example.com", &owner, "http://localhost:8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ShortURL != "http://localhost:8080/1c" {
		t.Errorf("expected short_url ending with '1c', got %q", resp.ShortURL)
	}
}

func TestResolve_CacheHit(t *testing.T) {
	cache := &mockCacheRepo{
		getFunc: func(code string) (string, bool) {
			return "https://example.com", true
		},
	}
	svc := NewShortenerService(nil, cache)
	url, err := svc.Resolve("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://example.com" {
		t.Errorf("got %q, want %q", url, "https://example.com")
	}
}

func TestResolve_CacheMiss_DBHit(t *testing.T) {
	linkRepo := &mockLinkRepo{
		findByCode: func(code string) (*model.Link, error) {
			return &model.Link{OriginalURL: "https://example.com"}, nil
		},
	}
	cache := &mockCacheRepo{
		getFunc: func(code string) (string, bool) { return "", false },
		setFunc: func(code, originalURL string, ttl int) error { return nil },
	}
	svc := NewShortenerService(linkRepo, cache)
	url, err := svc.Resolve("abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://example.com" {
		t.Errorf("got %q, want %q", url, "https://example.com")
	}
}

func TestResolve_NotFound(t *testing.T) {
	linkRepo := &mockLinkRepo{
		findByCode: func(code string) (*model.Link, error) {
			return nil, nil
		},
	}
	cache := &mockCacheRepo{
		getFunc: func(code string) (string, bool) { return "", false },
	}
	svc := NewShortenerService(linkRepo, cache)
	_, err := svc.Resolve("nonexistent")
	if !errors.Is(err, ErrLinkNotFound) {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

func TestResolve_AfterCacheSet(t *testing.T) {
	var cachedCode, cachedURL string
	linkRepo := &mockLinkRepo{
		findByCode: func(code string) (*model.Link, error) {
			return &model.Link{OriginalURL: "https://example.com"}, nil
		},
	}
	cache := &mockCacheRepo{
		getFunc: func(code string) (string, bool) { return "", false },
		setFunc: func(code, originalURL string, ttl int) error {
			cachedCode = code
			cachedURL = originalURL
			return nil
		},
	}
	svc := NewShortenerService(linkRepo, cache)
	svc.Resolve("abc123")
	if cachedCode != "abc123" || cachedURL != "https://example.com" {
		t.Errorf("cache.Set not called properly: %q, %q", cachedCode, cachedURL)
	}
}

func TestGetStats(t *testing.T) {
	linkRepo := &mockLinkRepo{
		statsFunc: func(linkID int64) (*model.LinkStats, error) {
			return &model.LinkStats{TotalClicks: 42}, nil
		},
	}
	svc := NewShortenerService(linkRepo, nil)
	stats, err := svc.GetStats(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalClicks != 42 {
		t.Errorf("expected 42 clicks, got %d", stats.TotalClicks)
	}
}
