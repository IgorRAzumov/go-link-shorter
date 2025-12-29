CREATE TABLE IF NOT EXISTS links
(
    uuid
              VARCHAR(36) PRIMARY KEY,
    short_key VARCHAR(255) NOT NULL UNIQUE,
    full_url  TEXT         NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_links_short_key ON links (short_key);
CREATE INDEX IF NOT EXISTS idx_links_full_url ON links (full_url);

