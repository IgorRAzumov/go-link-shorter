ALTER TABLE links
    ADD COLUMN IF NOT EXISTS user_id VARCHAR(36);

CREATE INDEX IF NOT EXISTS idx_links_user_id ON links (user_id);
