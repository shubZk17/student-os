-- Refresh token lookups happen on every token refresh/logout.
CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens(token_hash);

-- Emails are compared case-insensitively. Fails if case-duplicate accounts already exist;
-- merge those before applying.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users (lower(email));

ALTER TABLE applications DROP CONSTRAINT IF EXISTS chk_applications_stage;
ALTER TABLE applications ADD CONSTRAINT chk_applications_stage
    CHECK (stage IN ('SAVED', 'APPLIED', 'ASSESSMENT', 'INTERVIEW', 'OFFER', 'REJECTED', 'WITHDRAWN'));
