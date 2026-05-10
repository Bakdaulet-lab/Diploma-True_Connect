-- Add IIN hash uniqueness constraint for KYC ban-evasion detection (E1)
ALTER TABLE identity_vault.verifications
ADD COLUMN iin_hash BYTEA NOT NULL DEFAULT ''::bytea;

CREATE UNIQUE INDEX idx_verifications_iin_hash
ON identity_vault.verifications(iin_hash)
WHERE iin_hash != ''::bytea;
