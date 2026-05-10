DROP INDEX IF EXISTS identity_vault.idx_verifications_iin_hash;
ALTER TABLE identity_vault.verifications
DROP COLUMN iin_hash;
