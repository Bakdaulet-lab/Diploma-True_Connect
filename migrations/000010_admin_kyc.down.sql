DROP INDEX IF EXISTS identity_vault.idx_verifications_status;
ALTER TABLE identity_vault.verifications DROP COLUMN document_url;
ALTER TABLE identity_vault.verifications DROP COLUMN status;
