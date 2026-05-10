-- 000013_halal_identity_fields.down.sql

ALTER TABLE social.user_settings
    DROP COLUMN IF EXISTS modesty_level,
    DROP COLUMN IF EXISTS niyyah_filter,
    DROP COLUMN IF EXISTS madhab_filter;

DROP INDEX IF EXISTS idx_profiles_niyyah;

ALTER TABLE social.profiles
    DROP COLUMN IF EXISTS niyyah,
    DROP COLUMN IF EXISTS madhab,
    DROP COLUMN IF EXISTS languages,
    DROP COLUMN IF EXISTS no_photo_mode;

DROP TYPE IF EXISTS social.madhab;
DROP TYPE IF EXISTS social.niyyah;
