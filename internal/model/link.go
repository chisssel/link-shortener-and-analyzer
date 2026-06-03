package model

import "time"

type Link struct {
	ID          int64      `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	OwnerID     *string    `json:"owner_id,omitempty"`
}

type Click struct {
	ID         int64     `json:"id"`
	LinkID     int64     `json:"link_id"`
	ClickedAt  time.Time `json:"clicked_at"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Referer    string    `json:"referer"`
	Country    string    `json:"country"`
	City       string    `json:"city"`
}

type CreateLinkRequest struct {
	OriginalURL string  `json:"original_url" binding:"required"`
	OwnerID     *string `json:"owner_id,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type CreateLinkResponse struct {
	ShortURL    string `json:"short_url"`
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
}

type LinkStats struct {
	TotalClicks  int64          `json:"total_clicks"`
	UniqueClicks int64          `json:"unique_clicks"`
	ClicksByDay  []ClickByDay   `json:"clicks_by_day"`
	TopReferrers []ReferrerStat `json:"top_referrers"`
	TopCountries []CountryStat  `json:"top_countries"`
}

type ClickByDay struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Count    int64  `json:"count"`
}

type CountryStat struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

type LinkRepository interface {
	Insert(originalURL string, ownerID *string) (*Link, error)
	UpdateShortCode(id int64, shortCode string) error
	FindByShortCode(code string) (*Link, error)
	FindByID(id int64) (*Link, error)
	FindByOwner(ownerID string) ([]Link, error)
	Delete(id int64) error
	InsertClick(click *Click) error
	GetStats(linkID int64) (*LinkStats, error)
}

type CacheRepository interface {
	Get(url string) (string, bool)
	Set(code, originalURL string, ttl int) error
	DeleteCache(code string) error
}
