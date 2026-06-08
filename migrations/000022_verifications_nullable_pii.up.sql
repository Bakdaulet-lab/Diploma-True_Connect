-- Photo-based KYC does not collect IIN/document hash/full name.
-- Make legacy Sumsub-era columns nullable so SubmitKYCRequest can insert
-- without them.
ALTER TABLE identity_vault.verifications
    ALTER COLUMN iin_encrypted       DROP NOT NULL,
    ALTER COLUMN document_type       DROP NOT NULL,
    ALTER COLUMN document_hash       DROP NOT NULL,
    ALTER COLUMN full_name_encrypted DROP NOT NULL,
    ALTER COLUMN verification_method DROP NOT NULL;
