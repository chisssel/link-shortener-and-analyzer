CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE IF NOT EXISTS links (
    id          BIGSERIAL PRIMARY KEY,
    short_code  VARCHAR(10) NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ,
    owner_id    UUID
);

CREATE INDEX idx_links_short_code ON links (short_code);
CREATE INDEX idx_links_owner_id ON links (owner_id);

CREATE TABLE IF NOT EXISTS clicks (
    id          BIGSERIAL PRIMARY KEY,
    link_id     BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    clicked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address  INET,
    user_agent  TEXT,
    referer     TEXT,
    country     VARCHAR(2),
    city        VARCHAR(100)
);

CREATE INDEX idx_clicks_link_id ON clicks (link_id);
CREATE INDEX idx_clicks_clicked_at ON clicks (clicked_at);
