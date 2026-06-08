ALTER TABLE identity_vault.verifications
    ALTER COLUMN iin_encrypted       SET NOT NULL,
    ALTER COLUMN document_type       SET NOT NULL,
    ALTER COLUMN document_hash       SET NOT NULL,
    ALTER COLUMN full_name_encrypted SET NOT NULL,
    ALTER COLUMN verification_method SET NOT NULL;
