-- 000014_mahram.up.sql
-- Mahram (chaperone) registration for female users

CREATE TABLE social.mahrams (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    woman_user_id           UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    mahram_phone_encrypted  BYTEA NOT NULL,
    mahram_phone_hash       BYTEA NOT NULL UNIQUE,
    telegram_chat_id        BIGINT,
    verification_status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    verified_at             TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mahrams_woman_user_id ON social.mahrams (woman_user_id);
