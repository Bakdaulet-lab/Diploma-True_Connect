-- 000001_init_schema.up.sql
-- Core schema, extensions, and enums for TrueConnect

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- SCHEMA: social
-- ============================================================
CREATE SCHEMA IF NOT EXISTS social;

CREATE TYPE social.verification_level AS ENUM (
    'none',
    'phone_verified',
    'id_verified',
    'photo_verified'
);

CREATE TYPE social.trust_status AS ENUM (
    'normal',
    'under_review',
    'suspended',
    'banned'
);

CREATE TYPE social.gender AS ENUM ('male', 'female', 'other');

-- USERS
CREATE TABLE social.users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone_hash      BYTEA NOT NULL UNIQUE,
    phone_encrypted BYTEA NOT NULL,
    email_encrypted BYTEA,
    password_hash   TEXT NOT NULL,
    verification_level social.verification_level NOT NULL DEFAULT 'none',
    trust_status    social.trust_status NOT NULL DEFAULT 'normal',
    trust_score     SMALLINT NOT NULL DEFAULT 50 CHECK (trust_score BETWEEN 0 AND 100),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- PROFILES
CREATE TABLE social.profiles (
    user_id         UUID PRIMARY KEY REFERENCES social.users(id) ON DELETE CASCADE,
    display_name    VARCHAR(60) NOT NULL,
    bio             VARCHAR(500),
    gender          social.gender,
    birth_date      DATE,
    city            VARCHAR(100),
    location        GEOGRAPHY(POINT, 4326),
    looking_for     social.gender,
    avatar_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_location ON social.profiles USING GIST (location);
CREATE INDEX idx_profiles_city ON social.profiles (city);

-- MEDIA
CREATE TABLE social.media (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    object_key      TEXT NOT NULL,
    media_type      VARCHAR(20) NOT NULL DEFAULT 'photo',
    sort_order      SMALLINT NOT NULL DEFAULT 0,
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_user ON social.media (user_id, sort_order);

-- POSTS
CREATE TABLE social.posts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author_id       UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    content         TEXT NOT NULL CHECK (char_length(content) BETWEEN 1 AND 2000),
    media_url       TEXT,
    like_count      INTEGER NOT NULL DEFAULT 0,
    comment_count   INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_author ON social.posts (author_id, created_at DESC);
CREATE INDEX idx_posts_feed ON social.posts (created_at DESC);

-- POST_LIKES
CREATE TABLE social.post_likes (
    post_id         UUID NOT NULL REFERENCES social.posts(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

-- POST_COMMENTS
CREATE TABLE social.post_comments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id         UUID NOT NULL REFERENCES social.posts(id) ON DELETE CASCADE,
    author_id       UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    content         TEXT NOT NULL CHECK (char_length(content) BETWEEN 1 AND 500),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_post ON social.post_comments (post_id, created_at);

-- INTERACTIONS
CREATE TABLE social.interactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rater_id        UUID NOT NULL REFERENCES social.users(id),
    rated_id        UUID NOT NULL REFERENCES social.users(id),
    rating          SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    context         VARCHAR(30) NOT NULL,
    comment         VARCHAR(300),
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT no_self_rating CHECK (rater_id != rated_id)
);

CREATE INDEX idx_interactions_rated ON social.interactions (rated_id, created_at DESC);
CREATE INDEX idx_interactions_rater ON social.interactions (rater_id, created_at DESC);

-- MATCHES
CREATE TABLE social.matches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_a_id       UUID NOT NULL REFERENCES social.users(id),
    user_b_id       UUID NOT NULL REFERENCES social.users(id),
    user_a_liked    BOOLEAN NOT NULL DEFAULT FALSE,
    user_b_liked    BOOLEAN NOT NULL DEFAULT FALSE,
    matched_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ordered_pair CHECK (user_a_id < user_b_id),
    CONSTRAINT unique_pair UNIQUE (user_a_id, user_b_id)
);

-- MESSAGES
CREATE TABLE social.messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    match_id        UUID NOT NULL REFERENCES social.matches(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL REFERENCES social.users(id),
    content_encrypted BYTEA NOT NULL,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_match ON social.messages (match_id, created_at);

-- REPORTS
CREATE TABLE social.reports (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id     UUID NOT NULL REFERENCES social.users(id),
    reported_id     UUID NOT NULL REFERENCES social.users(id),
    reason          VARCHAR(50) NOT NULL,
    description     VARCHAR(500),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at     TIMESTAMPTZ
);

-- USER_SETTINGS
CREATE TABLE social.user_settings (
    user_id             UUID PRIMARY KEY REFERENCES social.users(id) ON DELETE CASCADE,
    push_notifications  BOOLEAN NOT NULL DEFAULT TRUE,
    show_online_status  BOOLEAN NOT NULL DEFAULT TRUE,
    distance_unit       VARCHAR(5) NOT NULL DEFAULT 'km',
    max_distance_km     SMALLINT NOT NULL DEFAULT 50,
    age_range_min       SMALLINT NOT NULL DEFAULT 18,
    age_range_max       SMALLINT NOT NULL DEFAULT 60,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- REFRESH_TOKENS
CREATE TABLE social.refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    token_hash      BYTEA NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON social.refresh_tokens (user_id);

-- ============================================================
-- SCHEMA: identity_vault (restricted access)
-- ============================================================
CREATE SCHEMA IF NOT EXISTS identity_vault;

CREATE TABLE identity_vault.verifications (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL UNIQUE REFERENCES social.users(id) ON DELETE CASCADE,
    iin_encrypted       BYTEA NOT NULL,
    document_type       VARCHAR(30) NOT NULL,
    document_hash       BYTEA NOT NULL,
    full_name_encrypted BYTEA NOT NULL,
    verification_method VARCHAR(30) NOT NULL,
    verified_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE identity_vault.access_log (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,
    accessed_by     VARCHAR(100) NOT NULL,
    action          VARCHAR(30) NOT NULL,
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vault_access_log ON identity_vault.access_log (user_id, created_at DESC);
