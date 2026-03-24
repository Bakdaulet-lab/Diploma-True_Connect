-- 000001_init_schema.down.sql
-- Reverse everything from the up migration

DROP TABLE IF EXISTS identity_vault.access_log;
DROP TABLE IF EXISTS identity_vault.verifications;
DROP SCHEMA IF EXISTS identity_vault;

DROP INDEX IF EXISTS social.idx_refresh_tokens_user;
DROP TABLE IF EXISTS social.refresh_tokens;
DROP TABLE IF EXISTS social.user_settings;
DROP TABLE IF EXISTS social.reports;
DROP INDEX IF EXISTS social.idx_messages_match;
DROP TABLE IF EXISTS social.messages;
DROP TABLE IF EXISTS social.matches;
DROP INDEX IF EXISTS social.idx_interactions_rater;
DROP INDEX IF EXISTS social.idx_interactions_rated;
DROP TABLE IF EXISTS social.interactions;
DROP INDEX IF EXISTS social.idx_comments_post;
DROP TABLE IF EXISTS social.post_comments;
DROP TABLE IF EXISTS social.post_likes;
DROP INDEX IF EXISTS social.idx_posts_feed;
DROP INDEX IF EXISTS social.idx_posts_author;
DROP TABLE IF EXISTS social.posts;
DROP INDEX IF EXISTS social.idx_media_user;
DROP TABLE IF EXISTS social.media;
DROP INDEX IF EXISTS social.idx_profiles_city;
DROP INDEX IF EXISTS social.idx_profiles_location;
DROP TABLE IF EXISTS social.profiles;
DROP TABLE IF EXISTS social.users;

DROP TYPE IF EXISTS social.gender;
DROP TYPE IF EXISTS social.trust_status;
DROP TYPE IF EXISTS social.verification_level;
DROP SCHEMA IF EXISTS social;
