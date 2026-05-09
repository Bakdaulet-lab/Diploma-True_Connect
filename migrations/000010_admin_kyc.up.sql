ALTER TABLE identity_vault.verifications ADD COLUMN status VARCHAR(20) DEFAULT 'pending';
ALTER TABLE identity_vault.verifications ADD COLUMN document_url VARCHAR(255);
CREATE INDEX idx_verifications_status ON identity_vault.verifications(status);
