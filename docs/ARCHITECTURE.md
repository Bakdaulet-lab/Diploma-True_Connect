# TrueConnect — Master Development Plan & Architecture Document

**Version:** 1.0
**Date:** 2026-03-16
**Author:** Architecture Blueprint (Principal Engineer Review)
**Status:** DRAFT — Awaiting Founder Approval

---

## Table of Contents

1. [High-Level System Architecture](#1-high-level-system-architecture)
2. [Database Architecture & Schema Design](#2-database-architecture--schema-design)
3. [API & Communication Strategy](#3-api--communication-strategy)
4. [Development Roadmap](#4-development-roadmap)
5. [AI-Assisted Coding & Security Rules](#5-ai-assisted-coding--security-rules)

---

## 1. High-Level System Architecture

### 1.1 Component Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                        CLIENTS                                      │
│   ┌──────────────┐   ┌──────────────┐   ┌──────────────────────┐   │
│   │  Flutter App  │   │  Next.js Web │   │  Admin Dashboard     │   │
│   │  (iOS/Android)│   │ [DEPRECATED] │   │  (Next.js internal)  │   │
│   │  PRIMARY     │   │  legacy v6   │   │  (internal only)     │   │
│   └──────┬───────┘   └──────┬───────┘   └──────────┬───────────┘   │
└──────────┼──────────────────┼──────────────────────┼───────────────┘
           │                  │                      │
           ▼                  ▼                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     NGINX REVERSE PROXY                             │
│         TLS termination · Rate limiting · Static assets             │
│         Route: /api/* → Go backend · /* → Next.js                   │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
           ┌───────────────────┼───────────────────┐
           ▼                   ▼                   ▼
┌──────────────────┐ ┌─────────────────┐ ┌─────────────────────────┐
│  AUTH SERVICE    │ │  CORE API       │ │  REALTIME SERVICE       │
│  (Go / Gin)     │ │  (Go / Gin)     │ │  (Go / gorilla/websocket│
│                  │ │                  │ │   or nhooyr/websocket)  │
│ • Registration   │ │ • Profiles      │ │                         │
│ • Login/Logout   │ │ • Posts/Feed    │ │ • Chat messaging        │
│ • JWT issuance   │ │ • Matching      │ │ • Reputation push       │
│ • Refresh tokens │ │ • Reputation    │ │ • Presence/typing       │
│ • KYC trigger    │ │ • Reporting     │ │                         │
└────────┬─────────┘ └───────┬─────────┘ └───────────┬─────────────┘
         │                   │                       │
         ▼                   ▼                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     INTERNAL SERVICE LAYER                          │
│                                                                     │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────────────┐   │
│  │ Trust Engine │  │ KYC/Identity │  │ Notification Service     │   │
│  │ (Go worker) │  │ Verifier     │  │ (Go — FCM/APNs push)    │   │
│  │             │  │ (Go + ext.   │  │                          │   │
│  │ Neo4j graph │  │  KYC API)    │  │                          │   │
│  │ queries     │  │              │  │                          │   │
│  └──────┬──────┘  └──────┬───────┘  └──────────────────────────┘   │
│         │                │                                          │
└─────────┼────────────────┼──────────────────────────────────────────┘
          │                │
          ▼                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     DATA LAYER                                      │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────┐  ┌───────────┐  │
│  │ PostgreSQL   │  │ Neo4j        │  │ Redis    │  │ MinIO     │  │
│  │              │  │              │  │          │  │ (S3)      │  │
│  │ • Users      │  │ • Trust Graph│  │ • Sessions│ │           │  │
│  │ • Profiles   │  │ • Sybil      │  │ • Cache  │  │ • Avatars │  │
│  │ • Posts      │  │   detection  │  │ • Rate   │  │ • Photos  │  │
│  │ • Messages   │  │ • Clusters   │  │   limits │  │ • KYC docs│  │
│  │ • KYC vault  │  │              │  │ • Rep.   │  │           │  │
│  │ • Settings   │  │              │  │   scores │  │           │  │
│  └──────────────┘  └──────────────┘  └──────────┘  └───────────┘  │
│                                                                     │
│  ALL deployed on KZ VPS — no foreign managed services for PII       │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.2 Data Flow — Key Paths

**Authentication Flow:**
1. Client sends `POST /api/v1/auth/register` or `/login` to Nginx.
2. Nginx forwards to Auth Service.
3. Auth Service validates input, hashes password (Argon2id), stores user row in PostgreSQL.
4. On login: Auth Service issues a short-lived JWT access token (15 min) in the response body and a long-lived refresh token (7 days) as an `HttpOnly; Secure; SameSite=Strict` cookie. Refresh token hash is stored in Redis (keyed by user ID) for revocation.
5. All subsequent API requests carry the JWT in the `Authorization: Bearer <token>` header. Gin middleware validates the JWT signature and expiry on every request.
6. On token refresh: client hits `POST /api/v1/auth/refresh`. Auth Service reads the cookie, verifies the hash against Redis, rotates the refresh token (one-time use), and issues a new access token.

**Reputation Calculation Flow:**
1. User A rates User B after a verified interaction → `POST /api/v1/interactions`.
2. Core API writes the interaction to PostgreSQL (audit log) and publishes an event to an internal Go channel (or Redis Pub/Sub for multi-instance).
3. Trust Engine worker consumes the event:
   a. Creates/updates `(A)-[:RATED {score, context, timestamp}]->(B)` edge in Neo4j.
   b. Runs a weighted PageRank variant on B's neighborhood to compute a new 0–100 score.
   c. Writes the computed score back to Redis (cache) and PostgreSQL (durable store).
4. Realtime Service picks up the score-change event and pushes it over WebSocket to User B's connected clients.

**Sybil Detection Flow (Async/Scheduled):**
1. A cron job (Go worker) triggers every N hours.
2. Queries Neo4j for clusters using community detection (Louvain algorithm via Neo4j GDS).
3. Flags accounts that form densely connected subgraphs with no external edges and low identity-verification levels.
4. Flagged accounts are marked `trust_status = 'under_review'` in PostgreSQL and excluded from matching.

### 1.3 Monorepo Directory Structure (Go Backend)

The backend is a **single monorepo** with internal packages following clean architecture. This keeps deployment simple for a small team while maintaining separation of concerns.

```
trueconnect/
├── cmd/                          # Application entry points
│   ├── api/                      # Main API server
│   │   └── main.go
│   ├── worker/                   # Background workers (trust engine, sybil cron)
│   │   └── main.go
│   └── migrate/                  # DB migration runner
│       └── main.go
│
├── internal/                     # Private application code (cannot be imported externally)
│   ├── config/                   # Configuration loading (env, files)
│   │   └── config.go
│   │
│   ├── domain/                   # Core business entities and interfaces (NO dependencies)
│   │   ├── user.go               # User, Profile, IdentityVault structs
│   │   ├── interaction.go        # Interaction, Rating structs
│   │   ├── post.go
│   │   ├── message.go
│   │   ├── reputation.go         # TrustScore value object
│   │   └── errors.go             # Domain-specific error types
│   │
│   ├── repository/               # Data access interfaces (ports)
│   │   ├── user_repo.go          # type UserRepository interface { ... }
│   │   ├── interaction_repo.go
│   │   ├── post_repo.go
│   │   ├── message_repo.go
│   │   └── trust_graph_repo.go   # Neo4j-facing interface
│   │
│   ├── service/                  # Business logic (use cases) — depends on domain + repository interfaces
│   │   ├── auth_service.go
│   │   ├── profile_service.go
│   │   ├── matching_service.go
│   │   ├── reputation_service.go
│   │   ├── interaction_service.go
│   │   ├── post_service.go
│   │   └── chat_service.go
│   │
│   ├── adapter/                  # Interface adapters (driven side — implements repository interfaces)
│   │   ├── postgres/
│   │   │   ├── user_repo.go      # implements repository.UserRepository
│   │   │   ├── post_repo.go
│   │   │   ├── message_repo.go
│   │   │   └── db.go             # Connection pool setup
│   │   ├── neo4j/
│   │   │   ├── trust_graph_repo.go
│   │   │   └── driver.go
│   │   ├── redis/
│   │   │   ├── session_store.go
│   │   │   ├── cache.go
│   │   │   └── rate_limiter.go
│   │   └── minio/
│   │       └── media_store.go
│   │
│   ├── handler/                  # HTTP handlers (driving side — Gin route handlers)
│   │   ├── auth_handler.go
│   │   ├── profile_handler.go
│   │   ├── post_handler.go
│   │   ├── interaction_handler.go
│   │   ├── matching_handler.go
│   │   ├── chat_handler.go       # WebSocket upgrade + handler
│   │   └── middleware/
│   │       ├── auth_middleware.go # JWT validation
│   │       ├── rate_limit.go
│   │       ├── cors.go
│   │       └── request_id.go
│   │
│   ├── worker/                   # Background job logic
│   │   ├── trust_engine.go       # Reputation recalculation
│   │   └── sybil_detector.go     # Cluster analysis cron
│   │
│   └── pkg/                      # Shared internal utilities
│       ├── crypto/               # Argon2 hashing, AES-256 encrypt/decrypt helpers
│       ├── validator/            # Input validation helpers
│       ├── jwt/                  # JWT creation and verification
│       └── logger/               # Structured logging (slog wrapper)
│
├── migrations/                   # SQL migration files (golang-migrate format)
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   └── ...
│
├── api/                          # API specification
│   └── openapi.yaml              # OpenAPI 3.1 spec
│
├── deployments/
│   ├── docker-compose.yml        # Local dev stack
│   ├── docker-compose.prod.yml   # Production overrides
│   ├── Dockerfile                # Multi-stage Go build
│   └── nginx/
│       └── nginx.conf
│
├── scripts/                      # Dev scripts (seed data, etc.)
├── .env.example
├── go.mod
├── go.sum
└── Makefile                      # build, test, lint, migrate commands
```

### 1.4 Key Architectural Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Monorepo vs Multi-repo | **Monorepo** (single Go module) | Solo/small team. Simplifies dependency management, CI, and refactoring. Split later if needed. |
| Monolith vs Microservices | **Modular monolith** | Single binary, internal package boundaries. Deploy as one process for MVP. Extract services later via the interface boundaries. |
| API Gateway | **Nginx** (not Kong/Traefik) | Lightweight, battle-tested TLS termination + reverse proxy. No service mesh needed at MVP scale. |
| ORM vs Raw SQL | **sqlc** (generates Go from SQL) | Type-safe, no reflection overhead, you write real SQL. Alternatives: pgx raw queries for complex cases. |
| Migration tool | **golang-migrate** | SQL-based, no DSL, explicit up/down files. |
| Neo4j driver | **neo4j/neo4j-go-driver** (official) | Maintained by Neo4j. Use Bolt protocol. |
| WebSocket library | **coder/websocket** (formerly nhooyr) | Standards-compliant, context-aware, lighter than gorilla (archived). |
| Structured logging | **log/slog** (stdlib, Go 1.21+) | Zero-dependency, JSON output, sufficient for MVP. |

---

## 2. Database Architecture & Schema Design

### 2.1 Privacy Architecture: Identity Vault Decoupling

The core privacy principle: **the public social profile never contains government identity data**. Two separate schemas achieve this:

```
┌─────────────────────────────────┐     ┌───────────────────────────────┐
│  SCHEMA: identity_vault         │     │  SCHEMA: social               │
│  (AES-256 encrypted at rest)    │     │  (public-facing data)         │
│                                  │     │                               │
│  ┌───────────────────────────┐  │     │  ┌─────────────────────────┐  │
│  │ identity_verifications    │  │     │  │ users                   │  │
│  │ • iin_encrypted (KZ ID)   │──┼──FK─┼─▶│ • id (UUID)             │  │
│  │ • document_type           │  │     │  │ • phone_hash            │  │
│  │ • verified_at             │  │     │  │ • email_encrypted       │  │
│  │ • verification_level      │  │     │  │ • verification_level    │  │
│  └───────────────────────────┘  │     │  └─────────────────────────┘  │
│                                  │     │                               │
│  Access: Auth Service ONLY       │     │  Access: All services          │
│  Audit-logged every read         │     │                               │
└─────────────────────────────────┘     └───────────────────────────────┘
```

- The `identity_vault` schema has a separate PostgreSQL role with restricted permissions.
- Application code accesses it only through a dedicated `IdentityRepository` interface.
- The `social` schema references identity only by `user_id` (UUID) and a denormalized `verification_level` enum.

### 2.2 PostgreSQL Schema (DDL)

```sql
-- ============================================================
-- EXTENSIONS
-- ============================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- SCHEMA: social (public-facing data)
-- ============================================================
CREATE SCHEMA IF NOT EXISTS social;

-- ENUM types
CREATE TYPE social.verification_level AS ENUM (
    'none',           -- unverified
    'phone_verified', -- SMS OTP confirmed
    'id_verified',    -- government ID confirmed via KYC
    'photo_verified'  -- liveness check passed
);

CREATE TYPE social.trust_status AS ENUM (
    'normal',
    'under_review',   -- flagged by Sybil detector
    'suspended',
    'banned'
);

CREATE TYPE social.gender AS ENUM ('male', 'female', 'other');

-- USERS — authentication and account-level data
CREATE TABLE social.users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone_hash      BYTEA NOT NULL UNIQUE,      -- SHA-256 of normalized phone (for lookup)
    phone_encrypted BYTEA NOT NULL,              -- AES-256-GCM encrypted phone (for display/contact)
    email_encrypted BYTEA,                       -- AES-256-GCM encrypted (optional)
    password_hash   TEXT NOT NULL,               -- Argon2id hash
    verification_level social.verification_level NOT NULL DEFAULT 'none',
    trust_status    social.trust_status NOT NULL DEFAULT 'normal',
    trust_score     SMALLINT NOT NULL DEFAULT 50 CHECK (trust_score BETWEEN 0 AND 100),
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- PROFILES — public social profile (decoupled from identity)
CREATE TABLE social.profiles (
    user_id         UUID PRIMARY KEY REFERENCES social.users(id) ON DELETE CASCADE,
    display_name    VARCHAR(60) NOT NULL,
    bio             VARCHAR(500),
    gender          social.gender,
    birth_date      DATE,                        -- age displayed, never exact date
    city            VARCHAR(100),
    location        GEOGRAPHY(POINT, 4326),      -- PostGIS geolocation for matching
    looking_for     social.gender,               -- matching preference
    avatar_url      TEXT,                         -- MinIO object path
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_location ON social.profiles USING GIST (location);
CREATE INDEX idx_profiles_city ON social.profiles (city);

-- MEDIA — user-uploaded photos/images
CREATE TABLE social.media (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    object_key      TEXT NOT NULL,                -- MinIO path: users/{user_id}/photos/{uuid}.webp
    media_type      VARCHAR(20) NOT NULL DEFAULT 'photo',  -- photo, video_thumbnail
    sort_order      SMALLINT NOT NULL DEFAULT 0,
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE, -- set TRUE after moderation
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_user ON social.media (user_id, sort_order);

-- POSTS — social feed content
CREATE TABLE social.posts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    author_id       UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    content         TEXT NOT NULL CHECK (char_length(content) BETWEEN 1 AND 2000),
    media_url       TEXT,                         -- optional attached image
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

-- INTERACTIONS — verified real-world meetings that feed the reputation system
CREATE TABLE social.interactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rater_id        UUID NOT NULL REFERENCES social.users(id),
    rated_id        UUID NOT NULL REFERENCES social.users(id),
    rating          SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    context         VARCHAR(30) NOT NULL,         -- 'date', 'meetup', 'event'
    comment         VARCHAR(300),
    is_verified     BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE if both parties confirmed the meeting
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT no_self_rating CHECK (rater_id != rated_id)
);

CREATE INDEX idx_interactions_rated ON social.interactions (rated_id, created_at DESC);
CREATE INDEX idx_interactions_rater ON social.interactions (rater_id, created_at DESC);

-- MATCHES — bidirectional like/match tracking
CREATE TABLE social.matches (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_a_id       UUID NOT NULL REFERENCES social.users(id),
    user_b_id       UUID NOT NULL REFERENCES social.users(id),
    user_a_liked    BOOLEAN NOT NULL DEFAULT FALSE,
    user_b_liked    BOOLEAN NOT NULL DEFAULT FALSE,
    matched_at      TIMESTAMPTZ,                  -- set when both liked
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ordered_pair CHECK (user_a_id < user_b_id),
    CONSTRAINT unique_pair UNIQUE (user_a_id, user_b_id)
);

-- MESSAGES — chat messages (only between matched users)
CREATE TABLE social.messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    match_id        UUID NOT NULL REFERENCES social.matches(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL REFERENCES social.users(id),
    content_encrypted BYTEA NOT NULL,             -- AES-256-GCM encrypted message body
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_match ON social.messages (match_id, created_at);

-- REPORTS — user-submitted abuse/fake reports
CREATE TABLE social.reports (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id     UUID NOT NULL REFERENCES social.users(id),
    reported_id     UUID NOT NULL REFERENCES social.users(id),
    reason          VARCHAR(50) NOT NULL,          -- 'fake_profile', 'harassment', 'scam', 'other'
    description     VARCHAR(500),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, reviewed, actioned, dismissed
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at     TIMESTAMPTZ
);

-- USER_SETTINGS — notification and privacy preferences
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

-- REFRESH_TOKENS — for JWT refresh token rotation
CREATE TABLE social.refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES social.users(id) ON DELETE CASCADE,
    token_hash      BYTEA NOT NULL UNIQUE,        -- SHA-256 of the refresh token
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
    iin_encrypted       BYTEA NOT NULL,           -- AES-256-GCM encrypted Kazakhstan IIN
    document_type       VARCHAR(30) NOT NULL,      -- 'kz_id_card', 'passport'
    document_hash       BYTEA NOT NULL,            -- SHA-256 hash for dedup (prevents re-registration)
    full_name_encrypted BYTEA NOT NULL,            -- AES-256-GCM encrypted
    verification_method VARCHAR(30) NOT NULL,       -- 'manual_review', 'egov_api', 'third_party_kyc'
    verified_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Audit log for every access to the identity vault
CREATE TABLE identity_vault.access_log (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL,
    accessed_by     VARCHAR(100) NOT NULL,         -- service name or admin user ID
    action          VARCHAR(30) NOT NULL,           -- 'read', 'write', 'verify'
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vault_access_log ON identity_vault.access_log (user_id, created_at DESC);

-- ============================================================
-- ROW-LEVEL SECURITY NOTE
-- ============================================================
-- In production, create a restricted PostgreSQL ROLE (e.g., vault_reader)
-- that ONLY the Auth Service uses. All other services connect with a role
-- that has NO SELECT permission on identity_vault schema.
--
-- ALTER DEFAULT PRIVILEGES IN SCHEMA identity_vault REVOKE ALL ON TABLES FROM public;
-- GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA identity_vault TO vault_reader;
```

### 2.3 Neo4j Graph Model

**Nodes:**

| Label | Properties | Description |
|---|---|---|
| `:User` | `uid` (UUID, indexed), `verification_level`, `trust_score`, `created_at` | Mirror of key user data for graph queries |

**Relationships (Edges):**

| Type | Direction | Properties | Description |
|---|---|---|---|
| `:RATED` | `(A)-[:RATED]->(B)` | `score` (1-5), `context`, `is_verified` (bool), `created_at` | User A rated User B after interaction |
| `:MET_WITH` | `(A)-[:MET_WITH]->(B)` | `confirmed_by_both` (bool), `created_at` | Real-world meeting (both confirmed) |
| `:REPORTED` | `(A)-[:REPORTED]->(B)` | `reason`, `created_at` | Abuse report edge |
| `:MATCHED` | `(A)-[:MATCHED]->(B)` | `created_at` | Mutual match |

**Key Cypher Queries:**

```cypher
// Create user node (on registration)
CREATE (u:User {uid: $uid, verification_level: 'none', trust_score: 50, created_at: datetime()})

// Add a rating edge
MATCH (a:User {uid: $rater_uid}), (b:User {uid: $rated_uid})
CREATE (a)-[:RATED {score: $score, context: $context, is_verified: $verified, created_at: datetime()}]->(b)

// Compute weighted trust score for a user
// Weight formula: verified interactions from high-trust users count more
MATCH (u:User {uid: $target_uid})<-[r:RATED]-(rater:User)
WHERE r.is_verified = true
WITH u, rater, r,
     CASE
       WHEN rater.verification_level IN ['id_verified', 'photo_verified'] THEN 1.5
       ELSE 1.0
     END AS id_weight,
     rater.trust_score / 100.0 AS trust_weight
WITH u,
     AVG(r.score * id_weight * trust_weight) AS weighted_avg,
     COUNT(r) AS rating_count
// Bayesian smoothing: blend toward 50 (neutral) with few ratings
WITH u,
     (weighted_avg * rating_count + 2.5 * 5) / (rating_count + 5) AS smoothed_score
SET u.trust_score = toInteger(smoothed_score * 20)  // scale 1-5 → 0-100
RETURN u.trust_score

// Sybil detection: find suspicious clusters (Louvain via GDS)
// Step 1: Project the graph
CALL gds.graph.project('trust-net', 'User', {
  RATED: {orientation: 'UNDIRECTED'},
  MET_WITH: {orientation: 'UNDIRECTED'}
})

// Step 2: Run community detection
CALL gds.louvain.stream('trust-net')
YIELD nodeId, communityId
WITH communityId, collect(gds.util.asNode(nodeId)) AS members, count(*) AS size
WHERE size >= 3
// Step 3: Check if cluster members have real external connections
UNWIND members AS m
OPTIONAL MATCH (m)-[:RATED|MET_WITH]-(external:User)
WHERE NOT external IN members
WITH communityId, size, count(external) AS external_connections, members
WHERE external_connections < size * 0.3  // fewer than 30% external edges → suspicious
RETURN communityId, size, external_connections,
       [m IN members | m.uid] AS suspect_uids
```

### 2.4 Redis Key Structures

```
# ──────────────────────────────────────────────────
# SESSION MANAGEMENT
# ──────────────────────────────────────────────────

# Active refresh token hash per user (for single-device or token rotation)
# Stored as a SET to support multi-device login
Key:    session:{user_id}:refresh_tokens
Type:   SET
Values: SHA-256 hashes of active refresh tokens
TTL:    7 days (matches refresh token expiry)

Example:
  SADD session:550e8400-e29b-41d4-a716-446655440000:refresh_tokens "a1b2c3d4..."
  EXPIRE session:550e8400-e29b-41d4-a716-446655440000:refresh_tokens 604800

# ──────────────────────────────────────────────────
# RATE LIMITING (sliding window counter)
# ──────────────────────────────────────────────────

# Per-IP rate limit (general API)
Key:    rl:ip:{ip_address}:{window_minute}
Type:   STRING (counter)
TTL:    120 seconds
Limit:  60 requests per minute

# Per-user rate limit (authenticated endpoints)
Key:    rl:user:{user_id}:{endpoint}:{window_minute}
Type:   STRING (counter)
TTL:    120 seconds
Limit:  Varies per endpoint

# Auth-specific brute force protection
Key:    rl:auth:fail:{phone_hash}
Type:   STRING (counter)
TTL:    900 seconds (15 min)
Limit:  5 failed attempts → temporary lockout

Example:
  INCR rl:ip:192.168.1.1:202603161430
  EXPIRE rl:ip:192.168.1.1:202603161430 120

# ──────────────────────────────────────────────────
# REPUTATION SCORE CACHE
# ──────────────────────────────────────────────────

# Cached trust score (avoids hitting Neo4j on every profile view)
Key:    trust:{user_id}
Type:   STRING (integer 0-100)
TTL:    600 seconds (10 min — refreshed on recalculation)

Example:
  SET trust:550e8400-e29b-41d4-a716-446655440000 72 EX 600

# ──────────────────────────────────────────────────
# ONLINE PRESENCE
# ──────────────────────────────────────────────────

# Track which users are currently online (for matching priority)
Key:    presence:{user_id}
Type:   STRING ("online")
TTL:    90 seconds (heartbeat-refreshed via WebSocket ping)

# ──────────────────────────────────────────────────
# MATCHING CACHE
# ──────────────────────────────────────────────────

# "Already seen" set to avoid showing the same profiles again
Key:    seen:{user_id}
Type:   SET (set of user_ids already shown)
TTL:    86400 seconds (24 hours, then reset)

# ──────────────────────────────────────────────────
# WEBSOCKET / PUBSUB CHANNELS
# ──────────────────────────────────────────────────

# Redis Pub/Sub channel for real-time events (chat, reputation updates)
Channel: ws:user:{user_id}
Purpose: Fan-out events to the correct WebSocket connection
         when running multiple API instances
```

---

## 3. API & Communication Strategy

### 3.1 RESTful API Design Conventions

**Base URL:** `https://api.trueconnect.kz/v1`

**Naming Rules:**
- All resource names are **lowercase, plural nouns**: `/users`, `/posts`, `/interactions`
- Nested resources for clear ownership: `/users/{id}/photos`, `/posts/{id}/comments`
- Actions that don't map to CRUD use a verb sub-resource: `/auth/login`, `/auth/refresh`, `/matches/{id}/confirm`
- Query parameters for filtering, pagination, sorting: `?page=1&per_page=20&sort=-created_at`
- Standard HTTP status codes: 200, 201, 204, 400, 401, 403, 404, 409, 422, 429, 500

**Response Envelope:**

```json
// Success
{
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 142
  }
}

// Error
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Phone number format is invalid",
    "details": [
      {"field": "phone", "reason": "must be a valid KZ phone number (+7XXXXXXXXXX)"}
    ]
  }
}
```

### 3.2 Endpoint Catalog

```
AUTH
  POST   /v1/auth/register          Register new user (phone + password)
  POST   /v1/auth/login             Login (returns JWT + sets refresh cookie)
  POST   /v1/auth/refresh           Rotate refresh token, issue new JWT
  POST   /v1/auth/logout            Revoke refresh token
  POST   /v1/auth/verify-phone      Submit SMS OTP code

USERS
  GET    /v1/users/me               Get current user's account data
  PATCH  /v1/users/me               Update account fields (email, password)
  DELETE /v1/users/me               Soft-delete account (GDPR-style)

PROFILES
  GET    /v1/profiles/{id}          View a user's public profile
  PUT    /v1/profiles/me            Create/update own profile
  GET    /v1/profiles/me/photos     List own photos
  POST   /v1/profiles/me/photos     Upload photo (multipart → MinIO)
  DELETE /v1/profiles/me/photos/{id} Remove photo

MATCHING
  GET    /v1/matching/candidates    Get next batch of profiles to swipe
  POST   /v1/matching/like          Like a user
  POST   /v1/matching/pass          Pass on a user
  GET    /v1/matches                List current matches

INTERACTIONS
  POST   /v1/interactions           Submit a rating after a meeting
  POST   /v1/interactions/{id}/confirm  Other party confirms the meeting happened
  GET    /v1/users/{id}/reputation  Get public trust score + rating summary

POSTS (Social Feed)
  GET    /v1/posts                  Feed (paginated, sorted by recency)
  POST   /v1/posts                  Create a post
  GET    /v1/posts/{id}             Get single post with comments
  DELETE /v1/posts/{id}             Delete own post
  POST   /v1/posts/{id}/like        Like a post
  DELETE /v1/posts/{id}/like        Unlike
  POST   /v1/posts/{id}/comments    Comment on a post

CHAT
  GET    /v1/matches/{id}/messages  Paginated message history (asc by created_at)
  WS     /v1/ws                     WebSocket connection (authenticated)

REPORTS
  POST   /v1/reports                Report a user

SETTINGS
  GET    /v1/settings               Get user preferences
  PATCH  /v1/settings               Update preferences

KYC (Identity Verification)
  POST   /v1/kyc/submit             Upload identity documents
  GET    /v1/kyc/status             Check verification status
```

### 3.3 Authentication Mechanism

```
┌──────────┐                          ┌──────────┐          ┌───────┐
│  Client  │                          │  Go API  │          │ Redis │
└────┬─────┘                          └────┬─────┘          └───┬───┘
     │                                     │                    │
     │  POST /auth/login                   │                    │
     │  {phone, password}                  │                    │
     │ ────────────────────────────────▶   │                    │
     │                                     │                    │
     │              Validate credentials   │                    │
     │              (Argon2id verify)       │                    │
     │                                     │                    │
     │              Generate JWT (15min)    │                    │
     │              Generate refresh token  │                    │
     │                                     │  SADD session:uid  │
     │                                     │ ──────────────────▶│
     │                                     │                    │
     │  200 OK                             │                    │
     │  Body: {access_token, expires_in}   │                    │
     │  Set-Cookie: refresh=<token>;       │                    │
     │    HttpOnly; Secure; SameSite=Strict│                    │
     │    Path=/v1/auth; Max-Age=604800    │                    │
     │ ◀────────────────────────────────   │                    │
     │                                     │                    │
     │  GET /v1/profiles/me                │                    │
     │  Authorization: Bearer <JWT>        │                    │
     │ ────────────────────────────────▶   │                    │
     │                                     │                    │
     │              Middleware: verify JWT  │                    │
     │              (signature + expiry)    │                    │
     │                                     │                    │
     │  200 OK {profile data}              │                    │
     │ ◀────────────────────────────────   │                    │
     │                                     │                    │
     │  ...15 min later, JWT expired...    │                    │
     │                                     │                    │
     │  POST /auth/refresh                 │                    │
     │  Cookie: refresh=<token>            │                    │
     │ ────────────────────────────────▶   │                    │
     │                                     │  SISMEMBER check   │
     │                                     │ ──────────────────▶│
     │                                     │                    │
     │              Rotate: delete old,    │  SREM old, SADD new│
     │              issue new refresh      │ ──────────────────▶│
     │              Issue new JWT          │                    │
     │                                     │                    │
     │  200 OK {new access_token}          │                    │
     │  Set-Cookie: refresh=<new_token>    │                    │
     │ ◀────────────────────────────────   │                    │
```

**JWT Claims:**

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "ver": "phone_verified",
  "tst": "normal",
  "iat": 1710590400,
  "exp": 1710591300
}
```

- `sub`: user UUID
- `ver`: verification level (avoids a DB lookup for basic authorization)
- `tst`: trust status (lets middleware block suspended users without hitting the DB)
- Short expiry (15 min) limits the damage window of a stolen token
- Refresh tokens are one-time-use (rotation). Reuse of an old refresh token triggers revocation of all tokens for that user (compromise detection).

### 3.4 WebSocket Protocol

**Connection:** Client opens `wss://api.trueconnect.kz/v1/ws` with the JWT in the first message (not query param, to avoid logging):

```json
// Client → Server (first message after connection)
{"type": "auth", "token": "<JWT>"}

// Server → Client (auth success)
{"type": "auth_ok", "user_id": "550e..."}

// Server → Client (auth failure)
{"type": "auth_error", "message": "token expired"}
```

**Message Types:**

```json
// Client sends a chat message
{"type": "chat_msg", "match_id": "...", "content": "Hello!"}

// Server delivers a chat message
{"type": "chat_msg", "match_id": "...", "sender_id": "...", "content": "Hello!", "id": "...", "created_at": "..."}

// Server pushes reputation update
{"type": "trust_update", "trust_score": 74}

// Typing indicator
{"type": "typing", "match_id": "...", "is_typing": true}

// Read receipt
{"type": "read", "match_id": "...", "last_read_msg_id": "..."}

// Server-side heartbeat (every 30s)
{"type": "ping"}
// Client responds
{"type": "pong"}
```

**Scaling with Redis Pub/Sub:** When multiple API instances run behind Nginx, a WebSocket connection lives on one instance. To deliver a message to User B who might be on a different instance:
1. Instance 1 receives the chat message from User A.
2. Instance 1 publishes to Redis channel `ws:user:{user_b_id}`.
3. Instance 2 (where User B's WebSocket lives) subscribes to that channel and forwards.

---

## 4. Development Roadmap

### Overview

| Sprint | Duration | Focus |
|---|---|---|
| Sprint 0 | 1 week | Project scaffolding, CI, Docker, DB setup |
| Sprint 1 | 2 weeks | Auth system, user registration, JWT flow |
| Sprint 2 | 2 weeks | Profiles, photo upload, basic matching |
| Sprint 3 | 2 weeks | Reputation system, Neo4j integration, interactions |
| Sprint 4 | 2 weeks | Real-time chat, feed, Sybil detection, polish |

---

### Sprint 0 — Foundation (Week 1)

**Goal:** Every developer (even if just you) can `docker compose up` and have the full stack running.

- [ ] Initialize Go module (`go mod init github.com/trueconnect/backend`)
- [ ] Set up the monorepo directory structure from Section 1.3
- [ ] Write the `Dockerfile` (multi-stage: build with `golang:1.23-alpine`, run with `alpine:3.19`)
- [ ] Write `docker-compose.yml` with services: `api`, `postgres`, `neo4j`, `redis`, `minio`, `nginx`
- [ ] Configure PostgreSQL: run all migration files from Section 2.2 via golang-migrate
- [ ] Configure Neo4j: constraints and indexes (`CREATE CONSTRAINT user_uid FOR (u:User) REQUIRE u.uid IS UNIQUE`)
- [ ] Set up MinIO: create default buckets (`avatars`, `photos`, `kyc-docs`)
- [ ] Write `Makefile` targets: `build`, `run`, `test`, `lint`, `migrate-up`, `migrate-down`
- [ ] Set up `slog` structured logger with JSON output
- [ ] Write the `internal/config/config.go` loader (reads from env vars, validates required fields)
- [ ] Basic Gin router skeleton with health check endpoint: `GET /v1/health → 200`
- [ ] Connect to PostgreSQL (pgx pool), Redis (go-redis), Neo4j (official driver), MinIO (minio-go)

**Deliverable:** Running `docker compose up && curl localhost/v1/health` returns `{"status": "ok"}`.

---

### Sprint 1 — Authentication & Users (Weeks 2–3)

**Goal:** Users can register, log in, refresh tokens, and log out. Phone verification is stubbed.

- [ ] Implement `internal/domain/user.go` — User, CreateUserInput structs
- [ ] Implement `internal/repository/user_repo.go` — UserRepository interface
- [ ] Implement `internal/adapter/postgres/user_repo.go` — pgx-backed implementation (sqlc or hand-written)
- [ ] Implement `internal/pkg/crypto/argon2.go` — Argon2id hash + verify
- [ ] Implement `internal/pkg/crypto/aes.go` — AES-256-GCM encrypt/decrypt helpers
- [ ] Implement `internal/pkg/jwt/jwt.go` — create + verify JWT tokens
- [ ] Implement `internal/service/auth_service.go`:
  - `Register(phone, password)` → hash password, encrypt phone, store user, create Neo4j node
  - `Login(phone, password)` → verify, issue JWT + refresh token
  - `Refresh(refreshToken)` → rotate, issue new pair
  - `Logout(refreshToken)` → revoke
- [ ] Implement `internal/adapter/redis/session_store.go` — refresh token hash storage
- [ ] Implement `internal/handler/auth_handler.go` — Gin handlers for `/auth/*` endpoints
- [ ] Implement `internal/handler/middleware/auth_middleware.go` — JWT validation middleware
- [ ] Implement `internal/handler/middleware/rate_limit.go` — Redis sliding window rate limiter
- [ ] Implement `internal/pkg/validator/` — phone number format validation (KZ: +7XXXXXXXXXX)
- [ ] Stub SMS verification: `POST /v1/auth/verify-phone` accepts any 6-digit code in dev mode
- [ ] Write integration tests: register → login → access protected route → refresh → logout → access fails
- [ ] Write unit tests for Argon2, AES, JWT packages

**Deliverable:** Full auth flow works via curl/Postman. Protected endpoints reject unauthenticated requests.

---

### Sprint 2 — Profiles & Matching (Weeks 4–5)

**Goal:** Users can create profiles, upload photos, and see potential matches based on location and preferences.

- [ ] Implement `internal/domain/profile.go`, `internal/domain/media.go`
- [ ] Implement profile repository + adapter (PostGIS queries for `ST_DWithin`)
- [ ] Implement `internal/service/profile_service.go`:
  - `CreateOrUpdateProfile(userID, input)` → upsert profile row
  - `GetProfile(userID)` → return profile with trust score from Redis cache
- [ ] Implement `internal/adapter/minio/media_store.go`:
  - `Upload(userID, file)` → store in MinIO, return object key
  - `GetPresignedURL(objectKey)` → generate time-limited access URL
- [ ] Implement photo upload endpoint with validation:
  - Max 8 photos per user
  - File type whitelist (JPEG, PNG, WebP)
  - Max file size 10MB
  - Resize to reasonable dimensions server-side (use `disintegration/imaging`)
- [ ] Implement `internal/service/matching_service.go`:
  - `GetCandidates(userID, limit)`:
    1. Read user preferences (gender, age range, max distance) from `user_settings`
    2. Query PostGIS: `ST_DWithin(location, $userLocation, $maxDistanceMeters)`
    3. Filter out already-seen users (check Redis `seen:{user_id}` set)
    4. Filter out blocked/reported users
    5. Sort by: trust_score DESC, last_login_at DESC (prioritize trusted, active users)
    6. Add returned user IDs to `seen:{user_id}` set in Redis
  - `Like(userID, targetID)` → update matches table, check for mutual match
  - `Pass(userID, targetID)` → add to seen set only
- [ ] Implement matching handler endpoints
- [ ] Implement user settings endpoints (GET + PATCH `/v1/settings`)
- [ ] Begin Flutter mobile client: auth screens (registration, login, OTP) and profile creation
- [ ] Begin Next.js web client: landing page, auth pages (SSR)

**Deliverable:** Users can create profiles, upload photos, and receive location-based match suggestions. Mutual likes create a match.

---

### Sprint 3 — Reputation System & Graph (Weeks 6–7)

**Goal:** Users can rate each other after meetings. Trust scores are computed via the Neo4j graph. Basic Sybil detection is operational.

- [ ] Implement `internal/domain/interaction.go`, `internal/domain/reputation.go`
- [ ] Implement `internal/repository/trust_graph_repo.go` — interface for Neo4j operations:
  - `AddRating(raterUID, ratedUID, score, context, verified)`
  - `AddMeeting(uidA, uidB, confirmedByBoth)`
  - `ComputeTrustScore(uid) → int`
  - `DetectSybilClusters() → []SuspectCluster`
- [ ] Implement `internal/adapter/neo4j/trust_graph_repo.go` — Cypher queries from Section 2.3
- [ ] Implement `internal/service/interaction_service.go`:
  - `SubmitRating(raterID, ratedID, rating, context)`:
    1. Validate: users must be matched (`matches` table)
    2. Rate limit: max 1 rating per pair per 7 days
    3. Write to PostgreSQL `interactions` table
    4. Write to Neo4j graph
    5. Trigger async trust score recomputation
  - `ConfirmInteraction(interactionID, userID)`:
    1. Mark `is_verified = true` in both PG and Neo4j
    2. Verified interactions have 2x weight in score calculation
- [ ] Implement `internal/service/reputation_service.go`:
  - `RecalculateScore(userID)`:
    1. Run weighted Cypher query (Section 2.3)
    2. Write new score to `social.users.trust_score` (PG) and `trust:{user_id}` (Redis)
    3. Update Neo4j node property
    4. Publish score-change event (for WebSocket push in Sprint 4)
  - `GetScore(userID)` → check Redis first, fallback to PG
- [ ] Implement `internal/worker/trust_engine.go`:
  - Listens on a Go channel or Redis Pub/Sub for `rating_created` events
  - Calls `ReputeService.RecalculateScore()`
- [ ] Implement `internal/worker/sybil_detector.go`:
  - Cron: runs every 6 hours
  - Calls `TrustGraphRepo.DetectSybilClusters()`
  - For each suspect cluster: set `trust_status = 'under_review'` in PG
  - Log findings for manual admin review
- [ ] Implement `cmd/worker/main.go` entry point (runs trust engine + sybil detector)
- [ ] Implement interaction and reputation handler endpoints
- [ ] Write tests: rating flow, score calculation with known graph topologies, Sybil detection with planted fake clusters

**Deliverable:** Reputation scores are live and update after each verified interaction. Sybil clusters are detected and flagged.

---

### Sprint 4 — Chat, Feed, KYC & Polish (Weeks 8–9)

**Goal:** Real-time chat between matches, social feed, KYC stub, and MVP hardening.

- [ ] Implement WebSocket infrastructure:
  - `internal/handler/chat_handler.go`:
    - Upgrade HTTP → WebSocket via `coder/websocket`
    - Authenticate via first message (JWT)
    - Maintain in-memory connection registry: `map[userID]*websocket.Conn`
    - Subscribe to Redis Pub/Sub channel `ws:user:{userID}` for cross-instance delivery
  - Message flow: Client → validate → encrypt content (AES-256) → store in PG → publish to Redis → deliver to recipient WebSocket
  - Typing indicators and read receipts (in-memory only, no persistence)
  - Heartbeat: server sends ping every 30s, disconnect on 3 missed pongs
- [ ] Implement `internal/service/chat_service.go`:
  - `SendMessage(senderID, matchID, content)` → validate match exists, encrypt, store, push
  - `GetMessages(matchID, page, perPage)` → decrypt and return
- [ ] Implement social feed:
  - `internal/service/post_service.go` — CRUD for posts, like/unlike, comment
  - `internal/handler/post_handler.go` — REST endpoints
  - Feed query: ordered by `created_at DESC`, paginated with cursor-based pagination
- [ ] KYC stub:
  - `POST /v1/kyc/submit` accepts file upload, stores in MinIO (`kyc-docs` bucket), sets `verification_level = 'pending'`
  - Manual review workflow: admin endpoint to approve/reject (sets `id_verified` or rejects)
  - On approval: update PG user, update Neo4j node, log in `identity_vault.access_log`
- [ ] Push notification stubs (FCM/APNs integration points — actual implementation deferred)
- [ ] Security hardening:
  - CORS configuration (whitelist trueconnect.kz origins only)
  - Helmet-style headers via Nginx (X-Content-Type-Options, X-Frame-Options, CSP)
  - Request ID middleware for traceability
  - Input sanitization on all text fields (strip HTML, limit lengths)
  - SQL injection: already mitigated by parameterized queries (sqlc/pgx), but audit all raw queries
- [ ] Flutter mobile: matching screen (swipe UI), chat screen, profile viewer, reputation display
- [ ] Next.js web: profile pages, feed, chat interface
- [ ] End-to-end integration tests: register → profile → match → chat → rate → reputation update
- [ ] Load testing: identify bottleneck under 100 concurrent WebSocket connections

**Deliverable:** Fully functional MVP — users can register, verify (stubbed), create profiles, match, chat in real time, post to a social feed, rate each other, and see trust scores update live.

---

## 5. AI-Assisted Coding & Security Rules

### 5.1 Go Coding Rules

**These rules MUST be followed in all future code generation prompts for this project.**

1. **Error handling is non-negotiable.** Every function that returns an error must have its error checked. Never use `_` to discard errors. Wrap errors with context using `fmt.Errorf("operation_name: %w", err)` for proper error chain unwrapping.

2. **Interface-driven design.** All data access is behind interfaces defined in `internal/repository/`. Business logic in `internal/service/` depends only on interfaces, never on concrete adapter implementations. This enables testing with mocks and swapping storage backends.

3. **Constructor injection.** Services receive their dependencies via constructor functions:
   ```go
   func NewAuthService(
       userRepo repository.UserRepository,
       sessionStore repository.SessionStore,
       cfg *config.AuthConfig,
   ) *AuthService
   ```

4. **Context propagation.** Every function that does I/O accepts `context.Context` as its first parameter. Use it for timeouts, cancellation, and request-scoped values (user ID, request ID).

5. **Struct validation.** Use a validation library (e.g., `go-playground/validator/v10`) on all input DTOs. Validate at the handler layer before passing to services.

6. **No global state.** No global DB connections, no `init()` hacks. Everything is wired in `main.go` and passed via constructors.

7. **Naming conventions:**
   - Package names: lowercase, single word (`handler`, `service`, `postgres`)
   - Interfaces: do NOT prefix with `I`. Use descriptive names: `UserRepository`, not `IUserRepository`
   - Unexported helpers: keep them in the same file as their public consumers
   - Methods on nil receivers: never design for this

8. **Testing:** Unit tests in `_test.go` files alongside the code. Integration tests in a `tests/` directory. Use `testify/assert` for assertions. Use table-driven tests wherever appropriate.

9. **SQL generation:** Prefer `sqlc` for query generation. For complex queries (CTEs, PostGIS), use raw `pgx` with parameterized queries ($1, $2 — NEVER string interpolation).

10. **Concurrency:** Use goroutines and channels for background work. Always enforce graceful shutdown (`signal.NotifyContext`). Use `errgroup` for parallel operations that must all succeed.

### 5.2 Security Rules

1. **Password hashing:** Argon2id with these parameters:
   - Memory: 64 MB
   - Iterations: 3
   - Parallelism: 4
   - Salt: 16 bytes (crypto/rand)
   - Key length: 32 bytes

2. **Encryption at rest:** AES-256-GCM for all PII fields (phone, email, IIN, full name, message content). Each encrypted field uses a unique nonce (12 bytes from crypto/rand). Encryption keys are loaded from environment variables, never hardcoded.

3. **SQL injection prevention:** Parameterized queries only. No string concatenation for SQL. sqlc enforces this by design. Any raw pgx query must use `$N` placeholders.

4. **Input validation:** All user input is validated and sanitized at the handler layer:
   - String length limits as defined in the schema
   - Phone: regex `^\+7\d{10}$`
   - Email: standard format validation
   - UUIDs: parsed via `uuid.Parse()`, never used raw
   - File uploads: type checked by magic bytes (not just extension), size limited

5. **Authentication:**
   - JWT signed with HS256 (symmetric). Key: 256-bit from env var. Migrate to RS256 (asymmetric) when adding third-party integrations.
   - Access token: 15-minute expiry, not stored anywhere server-side
   - Refresh token: 7-day expiry, stored as SHA-256 hash in Redis + PostgreSQL
   - One-time-use refresh tokens with rotation. Reuse detection triggers full session revocation.

6. **Authorization:**
   - Every mutating endpoint verifies `request.UserID == resource.OwnerID` (or has admin role)
   - Matching/chat: verify the users are actually matched before allowing messages
   - Rate interactions: verify the users have an existing match

7. **Transport security:**
   - TLS 1.2+ enforced at Nginx
   - HSTS header with 1-year max-age
   - No sensitive data in query parameters (tokens, passwords)
   - WebSocket auth via first message, not query parameter

8. **Rate limiting:** Applied at two layers:
   - Nginx: connection-level rate limiting (basic DDoS protection)
   - Go middleware: per-IP and per-user sliding window (Redis-backed)
   - Auth endpoints: aggressive limit (5 failures per 15 min per phone hash)

9. **Data sovereignty:** All PostgreSQL, Neo4j, Redis, and MinIO instances run on the KZ VPS. No replication to foreign datacenters. Backups are stored on the same KZ infrastructure. If external KYC APIs are used, only send the minimum required data and verify their data-processing jurisdiction.

10. **Logging:** Never log passwords, tokens, encrypted PII, or full request bodies containing sensitive data. Log: request IDs, user IDs, endpoints, status codes, latencies, error messages (sanitized).

### 5.3 Honesty Rule

**If I do not know the exact API, function signature, or Cypher query for a specific operation, I will explicitly state that I am unsure and provide a best-effort suggestion clearly marked as such. I will NOT invent package names, function signatures, or deprecated APIs. When recommending a third-party Go package, I will state the canonical import path and note whether I am confident it is currently maintained.**

---

## Appendix A: Recommended Go Dependencies

| Purpose | Package | Import Path | Notes |
|---|---|---|---|
| HTTP framework | Gin | `github.com/gin-gonic/gin` | Widely used, performant |
| PostgreSQL driver | pgx v5 | `github.com/jackc/pgx/v5` | Pure Go, production-grade |
| SQL code generation | sqlc | `github.com/sqlc-dev/sqlc` | CLI tool; generates Go from SQL |
| Migrations | golang-migrate | `github.com/golang-migrate/migrate/v4` | SQL-based up/down |
| Neo4j driver | Official | `github.com/neo4j/neo4j-go-driver/v5` | Bolt protocol |
| Redis client | go-redis v9 | `github.com/redis/go-redis/v9` | Context-aware, Pub/Sub support |
| MinIO client | minio-go v7 | `github.com/minio/minio-go/v7` | S3-compatible |
| WebSocket | coder/websocket | `github.com/coder/websocket` | Successor to nhooyr/websocket |
| Validation | validator v10 | `github.com/go-playground/validator/v10` | Struct tag validation |
| JWT | golang-jwt v5 | `github.com/golang-jwt/jwt/v5` | Actively maintained fork |
| UUID | google/uuid | `github.com/google/uuid` | V4 generation |
| Argon2 | stdlib | `golang.org/x/crypto/argon2` | Part of Go extended stdlib |
| Image processing | imaging | `github.com/disintegration/imaging` | Resize, crop |
| Testing | testify | `github.com/stretchr/testify` | Assertions + mocks |
| Env config | envconfig | `github.com/kelseyhightower/envconfig` | Struct-based env loading |
| Graceful shutdown | stdlib | `os/signal` + `context` | No external dep needed |

## Appendix B: Docker Compose Services (Reference)

```yaml
# docker-compose.yml (development)
services:
  api:
    build: .
    ports: ["8080:8080"]
    env_file: .env
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      neo4j:
        condition: service_healthy
      minio:
        condition: service_started

  postgres:
    image: postgis/postgis:16-3.4-alpine
    environment:
      POSTGRES_DB: trueconnect
      POSTGRES_USER: tc_app
      POSTGRES_PASSWORD: ${PG_PASSWORD}
    ports: ["5432:5432"]
    volumes:
      - pg_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U tc_app -d trueconnect"]
      interval: 5s
      retries: 5

  neo4j:
    image: neo4j:5-community
    environment:
      NEO4J_AUTH: neo4j/${NEO4J_PASSWORD}
      NEO4J_PLUGINS: '["graph-data-science"]'
    ports: ["7474:7474", "7687:7687"]
    volumes:
      - neo4j_data:/data
    healthcheck:
      test: ["CMD", "neo4j", "status"]
      interval: 10s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD} --maxmemory 256mb --maxmemory-policy allkeys-lru
    ports: ["6379:6379"]
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 5s
      retries: 5

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}
      MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY}
    ports: ["9000:9000", "9001:9001"]
    volumes:
      - minio_data:/data

  nginx:
    image: nginx:alpine
    ports: ["80:80", "443:443"]
    volumes:
      - ./deployments/nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - api

volumes:
  pg_data:
  neo4j_data:
  minio_data:
```

---

## Appendix C: Halal Pivot — New Database Tables (Sprints 6–8)

Added on branch `app-v7`. Migrations `000013`–`000016`. All tables in `social` schema unless noted.

### New ENUMs (migration 000013)

```sql
CREATE TYPE social.niyyah AS ENUM ('nikah_year', 'serious_marriage', 'friendship');
CREATE TYPE social.madhab AS ENUM ('hanafi', 'shafii', 'maliki', 'hanbali', 'none');
```

### Extended columns

| Table | New columns |
|---|---|
| `social.profiles` | `niyyah social.niyyah`, `madhab social.madhab`, `languages TEXT[]`, `no_photo_mode BOOLEAN DEFAULT false`, `marital_status VARCHAR(20) DEFAULT 'single'` |
| `social.user_settings` | `modesty_level INT DEFAULT 0`, `niyyah_filter TEXT`, `madhab_filter TEXT` |
| `social.matches` | `niyyah_timer_ends_at TIMESTAMPTZ`, `family_intro_done BOOLEAN DEFAULT false`, `imam_confirmed BOOLEAN DEFAULT false` |

### New tables (migrations 000014–000015)

```sql
-- Mahram (Islamic guardian) registration
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

-- 3-way mahram group chat rooms
CREATE TABLE social.mahram_chat_rooms (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    match_id        UUID UNIQUE REFERENCES social.matches(id),
    mahram_user_id  UUID REFERENCES social.users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Mahram group chat messages (AES-256-GCM encrypted)
CREATE TABLE social.mahram_messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id         UUID NOT NULL REFERENCES social.mahram_chat_rooms(id),
    sender_id       UUID NOT NULL REFERENCES social.users(id),
    content_encrypted BYTEA NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Whisper Network: anonymous post-match safety feedback
CREATE TABLE social.whisper_reports (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reporter_id             UUID NOT NULL REFERENCES social.users(id),
    reported_id             UUID NOT NULL REFERENCES social.users(id),
    meeting_match_id        UUID REFERENCES social.matches(id),
    feedback_encrypted      BYTEA NOT NULL,
    strike_weight           INT NOT NULL DEFAULT 1,
    strike_counted          BOOLEAN NOT NULL DEFAULT false,
    admin_flagged           BOOLEAN NOT NULL DEFAULT false,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### KYC IIN Uniqueness (migration 000016)

```sql
-- Prevents ban-evasion re-registration with same government ID
ALTER TABLE identity_vault.verifications
    ADD COLUMN iin_hash BYTEA NOT NULL DEFAULT ''::bytea;
CREATE UNIQUE INDEX idx_verifications_iin_hash
    ON identity_vault.verifications (iin_hash)
    WHERE iin_hash != ''::bytea;
```

### Frontend status (Sprint 8)

| Client | Status |
|---|---|
| Flutter (`frontend/`) | **Active** — primary mobile client (Sprints 12–13) |
| Next.js (`web/`) | **Deprecated** — frozen at v6, no longer maintained |

---

**End of Architecture Document v1.0**
