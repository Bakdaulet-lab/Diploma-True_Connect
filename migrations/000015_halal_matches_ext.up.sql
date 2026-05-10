-- 000015_halal_matches_ext.up.sql
-- Niyyah timer + family intro + imam confirmation + marital status
-- + mahram group chat rooms/messages + whisper reports

-- Match extensions
ALTER TABLE social.matches
    ADD COLUMN IF NOT EXISTS niyyah_timer_ends_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS family_intro_done     BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS imam_confirmed         BOOLEAN NOT NULL DEFAULT false;

-- Marital status on profiles (married via app → hidden from discovery)
ALTER TABLE social.profiles
    ADD COLUMN IF NOT EXISTS marital_status VARCHAR(20) NOT NULL DEFAULT 'single';

CREATE INDEX idx_profiles_marital_status ON social.profiles (marital_status);

-- Mahram group chat
CREATE TABLE social.mahram_chat_rooms (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    match_id        UUID NOT NULL UNIQUE REFERENCES social.matches(id) ON DELETE CASCADE,
    mahram_user_id  UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE social.mahram_messages (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id           UUID NOT NULL REFERENCES social.mahram_chat_rooms(id) ON DELETE CASCADE,
    sender_id         UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    content_encrypted BYTEA NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mahram_messages_room ON social.mahram_messages (room_id, created_at);

-- Whisper anonymous feedback network
CREATE TABLE social.whisper_reports (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id         UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    reported_id         UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    meeting_match_id    UUID REFERENCES social.matches(id) ON DELETE SET NULL,
    feedback_encrypted  BYTEA NOT NULL,
    strike_weight       INT NOT NULL DEFAULT 1,
    strike_counted      BOOLEAN NOT NULL DEFAULT false,
    admin_flagged       BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_whisper_reports_reported ON social.whisper_reports (reported_id, admin_flagged);
