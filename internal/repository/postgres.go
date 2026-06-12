package repository

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/artem/url-shortener/internal/model"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) Insert(originalURL string, ownerID *string) (*model.Link, error) {
	ctx := context.Background()
	var link model.Link
	err := r.pool.QueryRow(ctx,
		`INSERT INTO links (original_url, owner_id) VALUES ($1, $2)
		 RETURNING id, original_url, created_at, expires_at, owner_id`,
		originalURL, ownerID,
	).Scan(&link.ID, &link.OriginalURL, &link.CreatedAt, &link.ExpiresAt, &link.OwnerID)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *PostgresRepo) UpdateShortCode(id int64, shortCode string) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx,
		`UPDATE links SET short_code = $1 WHERE id = $2`, shortCode, id)
	return err
}

func (r *PostgresRepo) FindByShortCode(code string) (*model.Link, error) {
	ctx := context.Background()
	var link model.Link
	err := r.pool.QueryRow(ctx,
		`SELECT id, short_code, original_url, created_at, expires_at, owner_id
		 FROM links WHERE short_code = $1`, code,
	).Scan(&link.ID, &link.ShortCode, &link.OriginalURL, &link.CreatedAt, &link.ExpiresAt, &link.OwnerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		return nil, nil
	}
	return &link, nil
}

func (r *PostgresRepo) FindByID(id int64) (*model.Link, error) {
	ctx := context.Background()
	var link model.Link
	err := r.pool.QueryRow(ctx,
		`SELECT id, short_code, original_url, created_at, expires_at, owner_id
		 FROM links WHERE id = $1`, id,
	).Scan(&link.ID, &link.ShortCode, &link.OriginalURL, &link.CreatedAt, &link.ExpiresAt, &link.OwnerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &link, nil
}

func (r *PostgresRepo) FindByOwner(ownerID string) ([]model.Link, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx,
		`SELECT id, short_code, original_url, created_at, expires_at, owner_id
		 FROM links WHERE owner_id = $1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var links []model.Link
	for rows.Next() {
		var link model.Link
		if err := rows.Scan(&link.ID, &link.ShortCode, &link.OriginalURL, &link.CreatedAt, &link.ExpiresAt, &link.OwnerID); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func (r *PostgresRepo) Delete(id int64) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `DELETE FROM links WHERE id = $1`, id)
	return err
}

func (r *PostgresRepo) InsertClick(click *model.Click) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO clicks (link_id, ip_address, user_agent, referer, country, city)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		click.LinkID, parseIP(click.IPAddress), click.UserAgent, click.Referer, click.Country, click.City)
	return err
}

func (r *PostgresRepo) GetStats(linkID int64) (*model.LinkStats, error) {
	ctx := context.Background()
	stats := &model.LinkStats{}

	r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM clicks WHERE link_id = $1`, linkID).Scan(&stats.TotalClicks)

	r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT ip_address) FROM clicks WHERE link_id = $1`, linkID).Scan(&stats.UniqueClicks)

	rows, err := r.pool.Query(ctx,
		`SELECT DATE(clicked_at)::text, COUNT(*) FROM clicks
		 WHERE link_id = $1 GROUP BY 1 ORDER BY 1`, linkID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var d model.ClickByDay
			rows.Scan(&d.Date, &d.Count)
			stats.ClicksByDay = append(stats.ClicksByDay, d)
		}
	}

	rows, err = r.pool.Query(ctx,
		`SELECT COALESCE(NULLIF(referer, ''), 'direct'), COUNT(*) FROM clicks
		 WHERE link_id = $1 GROUP BY referer ORDER BY 2 DESC LIMIT 10`, linkID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r model.ReferrerStat
			rows.Scan(&r.Referrer, &r.Count)
			stats.TopReferrers = append(stats.TopReferrers, r)
		}
	}

	rows, err = r.pool.Query(ctx,
		`SELECT COALESCE(NULLIF(country, ''), 'unknown'), COUNT(*) FROM clicks
		 WHERE link_id = $1 GROUP BY country ORDER BY 2 DESC LIMIT 10`, linkID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c model.CountryStat
			rows.Scan(&c.Country, &c.Count)
			stats.TopCountries = append(stats.TopCountries, c)
		}
	}

	return stats, nil
}

func parseIP(addr string) net.IP {
	if strings.Contains(addr, ":") {
		host, _, err := net.SplitHostPort(addr)
		if err == nil {
			return net.ParseIP(host)
		}
	}
	return net.ParseIP(addr)
}
