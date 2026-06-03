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
	TotalClicks  int64            `json:"total_clicks"`
	UniqueClicks int64            `json:"unique_clicks"`
	ClicksByDay  []ClickByDay     `json:"clicks_by_day"`
	TopReferrers []ReferrerStat   `json:"top_referrers"`
	TopCountries []CountryStat    `json:"top_countries"`
}

type ClickByDay struct {
	Date   string `json:"date"`
	Count  int64  `json:"count"`
}

type ReferrerStat struct {
	Referrer string `json:"referrer"`
	Count    int64  `json:"count"`
}

type CountryStat struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}
