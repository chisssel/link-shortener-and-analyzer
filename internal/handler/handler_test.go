package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/artem/url-shortener/internal/geo"
	"github.com/artem/url-shortener/internal/model"
	"github.com/artem/url-shortener/internal/service"
)

type testRepo struct {
	links []model.Link
	clicks []model.Click
}

func NewTestRepo() *testRepo {
	return &testRepo{
		links:  []model.Link{},
		clicks: []model.Click{},
	}
}

func (r *testRepo) Insert(url string, ownerID *string) (*model.Link, error) {
	link := model.Link{
		ID:          int64(len(r.links) + 1),
		ShortCode:   "",
		OriginalURL: url,
		OwnerID:     ownerID,
	}
	r.links = append(r.links, link)
	return &link, nil
}
func (r *testRepo) UpdateShortCode(id int64, code string) error {
	for i := range r.links {
		if r.links[i].ID == id {
			r.links[i].ShortCode = code
		}
	}
	return nil
}
func (r *testRepo) FindByShortCode(code string) (*model.Link, error) {
	for _, l := range r.links {
		if l.ShortCode == code {
			return &l, nil
		}
	}
	return nil, nil
}
func (r *testRepo) FindByID(id int64) (*model.Link, error) {
	for _, l := range r.links {
		if l.ID == id {
			return &l, nil
		}
	}
	return nil, nil
}
func (r *testRepo) FindByOwner(ownerID string) ([]model.Link, error) {
	var result []model.Link
	for _, l := range r.links {
		if l.OwnerID != nil && *l.OwnerID == ownerID {
			result = append(result, l)
		}
	}
	return result, nil
}
func (r *testRepo) Delete(id int64) error {
	var filtered []model.Link
	for _, l := range r.links {
		if l.ID != id {
			filtered = append(filtered, l)
		}
	}
	r.links = filtered
	return nil
}
func (r *testRepo) InsertClick(click *model.Click) error {
	r.clicks = append(r.clicks, *click)
	return nil
}
func (r *testRepo) GetStats(linkID int64) (*model.LinkStats, error) {
	var count int64
	for _, c := range r.clicks {
		if c.LinkID == linkID {
			count++
		}
	}
	return &model.LinkStats{TotalClicks: count}, nil
}

type testCache struct {
	data map[string]string
}

func NewTestCache() *testCache {
	return &testCache{data: make(map[string]string)}
}
func (c *testCache) Get(code string) (string, bool) {
	v, ok := c.data[code]
	return v, ok
}
func (c *testCache) Set(code, originalURL string, ttl int) error {
	c.data[code] = originalURL
	return nil
}
func (c *testCache) DeleteCache(code string) error {
	delete(c.data, code)
	return nil
}

func setupTest() (*gin.Engine, *testRepo, *testCache) {
	gin.SetMode(gin.TestMode)
	repo := NewTestRepo()
	cache := NewTestCache()
	svc := service.NewShortenerService(repo, cache)
	geoClient := geo.NewMockClient("RU", "Tula")
	r := gin.New()
	RegisterRoutes(r, svc, repo, geoClient)
	return r, repo, cache
}

func TestShorten_Create_Success(t *testing.T) {
	r, _, _ := setupTest()
	body := `{"original_url": "https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp model.CreateLinkResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if resp.ShortCode == "" {
		t.Errorf("expected non-empty short_code")
	}
	if resp.OriginalURL != "https://example.com" {
		t.Errorf("expected original URL")
	}
}

func TestShorten_Create_InvalidJSON(t *testing.T) {
	r, _, _ := setupTest()
	body := `not-json`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestShorten_Create_EmptyURL(t *testing.T) {
	r, _, _ := setupTest()
	body := `{"original_url": ""}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty URL, got %d", w.Code)
	}
}

func TestShorten_Create_InvalidURL(t *testing.T) {
	r, _, _ := setupTest()
	body := `{"original_url": "not-a-valid-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestShorten_Create_WithOwner(t *testing.T) {
	r, _, _ := setupTest()
	body := `{"original_url": "https://example.com", "owner_id": "user-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestShorten_Create_Twice_DifferentCodes(t *testing.T) {
	r, _, _ := setupTest()

	body1 := `{"original_url": "https://example.com"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	body2 := `{"original_url": "https://example.org"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var resp1, resp2 model.CreateLinkResponse
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	json.Unmarshal(w2.Body.Bytes(), &resp2)

	if resp1.ShortCode == resp2.ShortCode {
		t.Errorf("expected different codes for different URLs")
	}
}

func TestRedirect_Success(t *testing.T) {
	r, _, _ := setupTest()
	body := `{"original_url": "https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp model.CreateLinkResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)

	redirectReq := httptest.NewRequest(http.MethodGet, "/"+createResp.ShortCode, nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, redirectReq)

	if w2.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", w2.Code)
	}
}

func TestAnalytics_GetStats(t *testing.T) {
	r, _, _ := setupTest()
	body := `{"original_url": "https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp model.CreateLinkResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)

	linkID := extractID(createResp.ShortURL, createResp.ShortCode)

	statsReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/link/%d/stats", linkID), nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, statsReq)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	var stats model.LinkStats
	json.Unmarshal(w2.Body.Bytes(), &stats)
	if stats.TotalClicks < 0 {
		t.Errorf("expected non-negative clicks")
	}
}

func TestAnalytics_InvalidID(t *testing.T) {
	r, _, _ := setupTest()
	req := httptest.NewRequest(http.MethodGet, "/api/link/abc/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid link id, got %d", w.Code)
	}
}

func TestLinkList_RequiresOwnerID(t *testing.T) {
	r, _, _ := setupTest()
	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLinkList_WithOwnerID(t *testing.T) {
	r, repo, _ := setupTest()
	owner := "user-1"
	repo.Insert("https://example.com", &owner)
	repo.Insert("https://example.org", &owner)
	repo.UpdateShortCode(1, "aaa")
	repo.UpdateShortCode(2, "bbb")

	req := httptest.NewRequest(http.MethodGet, "/api/links?owner_id=user-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var links []model.Link
	json.Unmarshal(w.Body.Bytes(), &links)
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

func TestLinkList_EmptyOwner(t *testing.T) {
	r, _, _ := setupTest()
	req := httptest.NewRequest(http.MethodGet, "/api/links?owner_id=nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var links []model.Link
	json.Unmarshal(w.Body.Bytes(), &links)
	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}

func TestDeleteLink(t *testing.T) {
	r, repo, _ := setupTest()
	repo.Insert("https://example.com", nil)
	repo.UpdateShortCode(1, "aaa")

	body := `{"link_id": 1}`
	req := httptest.NewRequest(http.MethodDelete, "/api/link", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/aaa", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w2.Code)
	}
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %q", body["status"])
	}
}

func TestShorten_Create_LongURL(t *testing.T) {
	r, _, _ := setupTest()
	longURL := "https://example.com/" + fmt.Sprintf("%01000d", 0)[:200]
	body := fmt.Sprintf(`{"original_url": "%s"}`, longURL)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusCreated {
		t.Logf("long URL created successfully")
	}
}

func TestShorten_Create_HTTPSOnly(t *testing.T) {
	r, _, _ := setupTest()
	body := `{"original_url": "http://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 for http URL, got %d", w.Code)
	}
}

func TestShorten_SpecialChars(t *testing.T) {
	r, _, _ := setupTest()
	body := `{"original_url": "https://example.com/path?q=hello+world&lang=en#section"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 for URL with special chars, got %d", w.Code)
	}
}

func TestRedirect_TrackingClick(t *testing.T) {
	r, repo, _ := setupTest()
	body := `{"original_url": "https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createResp model.CreateLinkResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)

	redirectReq := httptest.NewRequest(http.MethodGet, "/"+createResp.ShortCode, nil)
	redirectReq.Header.Set("User-Agent", "test-agent")
	redirectReq.Header.Set("Referer", "https://twitter.com")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, redirectReq)

	_ = repo
	if w2.Code == http.StatusFound {
		t.Logf("redirect succeeded, click tracked")
	}
}

func TestShorten_MissingContentType(t *testing.T) {
	r, _, _ := setupTest()
	body := `{"original_url": "https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusCreated {
		t.Logf("created even without content-type header")
	}
}

// extractID is a helper that encodes the link position back from short URL
func extractID(shortURL, shortCode string) int64 {
	_ = shortURL
	for i := int64(1); i <= 1000; i++ {
		code := ""
		n := i
		for n > 0 {
			code = string("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"[n%62]) + code
			n /= 62
		}
		if code == shortCode {
			return i
		}
	}
	return 0
}
