ALTER TABLE urls
    ADD COLUMN user_id TEXT NOT NULL;

CREATE INDEX idx_urls_user_id ON urls(user_id);

DROP INDEX IF EXISTS idx_unique_original_url;

CREATE UNIQUE INDEX idx_unique_original_user
    ON urls(original_url, user_id);