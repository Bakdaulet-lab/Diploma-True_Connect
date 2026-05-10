-- 000013_halal_identity_fields.up.sql
-- Islamic identity fields: niyyah/madhab ENUMs, profile + settings extensions

CREATE TYPE social.niyyah AS ENUM (
    'nikah_year',
    'serious_marriage',
    'friendship'
);

CREATE TYPE social.madhab AS ENUM (
    'hanafi',
    'shafii',
    'maliki',
    'hanbali',
    'none'
);

ALTER TABLE social.profiles
    ADD COLUMN IF NOT EXISTS niyyah      social.niyyah,
    ADD COLUMN IF NOT EXISTS madhab      social.madhab,
    ADD COLUMN IF NOT EXISTS languages   TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS no_photo_mode BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX idx_profiles_niyyah ON social.profiles (niyyah);

ALTER TABLE social.user_settings
    ADD COLUMN IF NOT EXISTS modesty_level  INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS niyyah_filter  TEXT,
    ADD COLUMN IF NOT EXISTS madhab_filter  TEXT;
