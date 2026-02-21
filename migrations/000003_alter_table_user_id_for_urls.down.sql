DROP INDEX IF EXISTS idx_unique_original_user;

CREATE UNIQUE INDEX idx_unique_original_url ON urls(original_url);

DROP INDEX IF EXISTS idx_urls_user_id;

ALTER TABLE urls
DROP COLUMN IF EXISTS user_id;