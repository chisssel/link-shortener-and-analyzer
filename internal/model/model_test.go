package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLink_Serialization(t *testing.T) {
	link := Link{
		ID:          1,
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
		CreatedAt:   time.Now(),
	}
	data, err := json.Marshal(link)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded Link
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.ID != 1 || decoded.ShortCode != "abc123" {
		t.Errorf("roundtrip failed: got %+v", decoded)
	}
}

func TestCreateLinkRequest_JSON(t *testing.T) {
	body := `{"original_url": "https://example.com"}`
	var req CreateLinkRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if req.OriginalURL != "https://example.com" {
		t.Errorf("expected https://example.com, got %q", req.OriginalURL)
	}
	if req.OwnerID != nil {
		t.Errorf("expected nil owner_id")
	}
}

func TestCreateLinkRequest_WithOwner(t *testing.T) {
	owner := "user-123"
	body := `{"original_url": "https://example.com", "owner_id": "user-123"}`
	var req CreateLinkRequest
	json.Unmarshal([]byte(body), &req)
	if req.OwnerID == nil || *req.OwnerID != owner {
		t.Errorf("expected owner_id %q, got %v", owner, req.OwnerID)
	}
}

func TestCreateLinkRequest_MissingURL(t *testing.T) {
	body := `{}`
	var req CreateLinkRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if req.OriginalURL != "" {
		t.Errorf("expected empty url")
	}
}

func TestCreateLinkResponse_JSON(t *testing.T) {
	resp := CreateLinkResponse{
		ShortURL:    "http://localhost:8080/abc123",
		ShortCode:   "abc123",
		OriginalURL: "https://example.com",
	}
	data, _ := json.Marshal(resp)
	var decoded CreateLinkResponse
	json.Unmarshal(data, &decoded)
	if decoded.ShortCode != "abc123" || decoded.ShortURL != "http://localhost:8080/abc123" {
		t.Errorf("roundtrip failed: %+v", decoded)
	}
}

func TestLinkStats_Empty(t *testing.T) {
	stats := LinkStats{}
	if stats.TotalClicks != 0 {
		t.Errorf("expected 0 total clicks, got %d", stats.TotalClicks)
	}
	if stats.ClicksByDay != nil {
		t.Errorf("expected nil clicks_by_day")
	}
}

func TestLinkStats_JSON(t *testing.T) {
	stats := LinkStats{
		TotalClicks:  100,
		UniqueClicks: 50,
		ClicksByDay: []ClickByDay{
			{Date: "2026-01-01", Count: 10},
			{Date: "2026-01-02", Count: 20},
		},
		TopReferrers: []ReferrerStat{
			{Referrer: "twitter.com", Count: 30},
		},
		TopCountries: []CountryStat{
			{Country: "RU", Count: 40},
		},
	}
	data, _ := json.Marshal(stats)
	var decoded LinkStats
	json.Unmarshal(data, &decoded)
	if decoded.TotalClicks != 100 || decoded.UniqueClicks != 50 {
		t.Errorf("basic fields mismatch: %+v", decoded)
	}
	if len(decoded.ClicksByDay) != 2 {
		t.Errorf("expected 2 days, got %d", len(decoded.ClicksByDay))
	}
	if len(decoded.TopReferrers) != 1 {
		t.Errorf("expected 1 referrer, got %d", len(decoded.TopReferrers))
	}
	if len(decoded.TopCountries) != 1 {
		t.Errorf("expected 1 country, got %d", len(decoded.TopCountries))
	}
}

func TestClick_Serialization(t *testing.T) {
	click := Click{
		LinkID:    1,
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Referer:   "https://google.com",
		Country:   "RU",
		City:      "Moscow",
	}
	data, _ := json.Marshal(click)
	var decoded Click
	json.Unmarshal(data, &decoded)
	if decoded.LinkID != 1 || decoded.IPAddress != "192.168.1.1" {
		t.Errorf("click fields mismatch: %+v", decoded)
	}
}
