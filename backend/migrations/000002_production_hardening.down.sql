ALTER TABLE IF EXISTS applications DROP CONSTRAINT IF EXISTS chk_applications_stage;
DROP INDEX IF EXISTS idx_users_email_lower;
DROP INDEX IF EXISTS idx_refresh_tokens_hash;
