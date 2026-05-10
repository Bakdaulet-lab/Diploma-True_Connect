-- 000015_halal_matches_ext.down.sql

DROP TABLE IF EXISTS social.whisper_reports;
DROP TABLE IF EXISTS social.mahram_messages;
DROP TABLE IF EXISTS social.mahram_chat_rooms;

DROP INDEX IF EXISTS idx_profiles_marital_status;
ALTER TABLE social.profiles
    DROP COLUMN IF EXISTS marital_status;

ALTER TABLE social.matches
    DROP COLUMN IF EXISTS niyyah_timer_ends_at,
    DROP COLUMN IF EXISTS family_intro_done,
    DROP COLUMN IF EXISTS imam_confirmed;
