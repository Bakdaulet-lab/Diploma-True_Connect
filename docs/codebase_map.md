# Codebase Knowledge Base — TrueConnect (Halal Dating App)

## Project Identity
- App name: **TrueConnect** (Казақша: Нағыз Байланыс)
- Go module: `github.com/trueconnect/backend`
- Backend language/framework: Go 1.25 + Gin
- Frontend: Flutter (Dart 3.2+), minimum Flutter SDK 3.16.0
- Databases: PostgreSQL 16 with PostGIS extension, Neo4j 5 Community, Redis 7
- Object storage: MinIO (S3-compatible, self-hosted)
- Authentication: JWT HS256 — 15-minute access tokens, 7-day refresh tokens (one-time-use rotation), Argon2id password hashing, AES-256-GCM for PII field encryption
- Target market: Kazakhstan and Central Asia (Muslim-majority population)

---

## Backend Architecture

- Architecture pattern: **Modular Monolith with Clean Architecture**
- Layer structure: `Handler → Service → Repository (interface) ← Adapter (implementation)`
- Domain isolation: `internal/domain/` contains only pure Go structs and error constants — zero external dependencies
- Inter-module communication: Go interfaces (ports); services depend only on repository interfaces, never on adapter packages
- Background processing: separate `cmd/worker` binary running TrustEngine and PushWorker goroutines

**Package structure:**
```
cmd/
  api/          ← main.go: wires everything together
  worker/       ← background workers
  migrate/      ← migration runner
internal/
  domain/       ← entities, enums, error constants
  repository/   ← interface definitions (ports)
  adapter/
    postgres/   ← PostgreSQL implementations
    neo4j/      ← Neo4j graph implementations
    redis/      ← Redis cache implementations
    minio/      ← MinIO object storage
    kyc/        ← KYC verification stub
  service/      ← business logic
  handler/      ← Gin HTTP handlers + WebSocket Hub
  worker/       ← TrustEngine, PushWorker, NiyyahTimerWorker
  pkg/
    crypto/     ← AES-256-GCM helpers
    jwt/        ← JWT sign/verify
    logger/     ← slog-based structured logging
    sanitize/   ← input sanitization
    validator/  ← input validation
    imam/       ← embedded imam catalog (//go:embed)
migrations/     ← golang-migrate SQL files (19 migrations)
deployments/    ← Dockerfile, docker-compose.yml, nginx/
frontend/       ← Flutter application
api/            ← openapi.yaml (49 paths)
```

---

## Backend Modules

### auth Module
- Handler: `internal/handler/auth_handler.go`
- Service: `internal/service/auth_service.go`
- Repository: `internal/adapter/postgres/user_repo.go`, `internal/adapter/postgres/refresh_token_repo.go`
- Key structs:
  - `User`: ID uuid, PhoneHash []byte, PhoneEncrypted []byte, EmailEncrypted []byte, PasswordHash string, PublicKey *string (X25519), VerificationLevel enum, TrustStatus enum, TrustScore int, IsAdmin bool, IsActive bool, LastLoginAt *time.Time, FCMToken *string, CreatedAt time.Time, UpdatedAt time.Time
  - `RefreshToken`: ID uuid, UserID uuid, TokenHash []byte, ExpiresAt time.Time, UsedAt *time.Time
- Enums: `VerificationLevel` (none, phone_verified, id_verified, photo_verified), `TrustStatus` (normal, under_review, suspended, banned)
- Key service functions:
  - `Register(ctx, phone, password, email) (*User, string, string, error)` — Argon2id hash, AES-256 encrypt PII, JWT pair
  - `Login(ctx, phone, password) (*User, string, string, error)` — Argon2id verify
  - `RefreshToken(ctx, rawToken) (string, string, error)` — one-time-use rotation
  - `Logout(ctx, userID, rawToken) error`
- API endpoints:
  - `POST /v1/auth/register` → RegisterHandler
  - `POST /v1/auth/login` → LoginHandler
  - `POST /v1/auth/refresh` → RefreshHandler (reads refresh token from HttpOnly cookie)
  - `POST /v1/auth/logout` → LogoutHandler
- DB tables: social.users, social.refresh_tokens
- Business rules: Argon2id (64MB memory, 3 iterations, 4 parallelism); refresh token family revocation on reuse detection; brute-force protection via per-IP Redis rate counter

---

### profile Module
- Handler: `internal/handler/profile_handler.go`
- Service: `internal/service/profile_service.go`
- Repository: `internal/adapter/postgres/profile_repo.go`, `internal/adapter/postgres/media_repo.go`, `internal/adapter/minio/media_store.go`
- Key structs:
  - `Profile`: UserID uuid, DisplayName string (max 60), Bio string (max 500), Gender enum, BirthDate *time.Time, City string (max 100), Latitude *float64, Longitude *float64, LookingFor Gender, AvatarURL string, Niyyah enum, Madhab enum, Languages []string, NoPhotoMode bool, MaritalStatus string, Prompts []PromptAnswer (JSONB), CreatedAt time.Time
  - `PromptAnswer`: Question string, Answer string
  - `Media`: ID uuid, UserID uuid, ObjectKey string, ContentType string, CreatedAt time.Time
- Enums: `Gender` (male, female, other), `Niyyah` (nikah_year, serious_marriage, friendship), `Madhab` (hanafi, shafii, maliki, hanbali, none)
- Key service functions:
  - `GetMyProfile(ctx, userID) (*Profile, error)`
  - `UpsertProfile(ctx, profile) error`
  - `UploadPhoto(ctx, userID, data []byte) (string, error)` — MinIO upload + DB record
  - `ComputeAge(birthDate time.Time) int` — exact year-aware calculation
  - `GetLeaderboard(ctx, limit int) ([]LeaderboardEntry, error)`
- API endpoints:
  - `GET /v1/profiles/me` → GetMyProfileHandler
  - `PATCH /v1/profiles/me` → UpdateProfileHandler
  - `GET /v1/profiles/:userId` → GetProfileHandler
  - `POST /v1/profiles/me/photos` → UploadPhotoHandler
  - `DELETE /v1/users/me` → DeleteAccountHandler (soft-delete + cascade)
- DB tables: social.profiles, social.media

---

### matching Module
- Handler: `internal/handler/matching_handler.go`
- Service: `internal/service/matching_service.go`
- Repository: `internal/adapter/postgres/match_repo.go`, `internal/adapter/postgres/profile_repo.go`
- Key structs:
  - `Match`: ID uuid, UserAID uuid, UserBID uuid, MatchedAt *time.Time, NiyyahTimerEndsAt *time.Time, FamilyIntroDone bool, ImamConfirmed bool
  - `LikeResult`: Matched bool, MatchID uuid
  - `CandidateRow`: UserID uuid, DisplayName string, AvatarURL string, TrustScore int, Niyyah string, Madhab string, Languages []string, NoPhotoMode bool
- Key service functions:
  - `GetCandidates(ctx, requesterID) ([]*CandidateCard, error)` — PostGIS distance + niyyah filter (B1) + madhab +10 boost (B2) + no-photo blur (B3) + seen-set exclusion
  - `Like(ctx, userID, targetID) (*LikeResult, error)` — mutual check, match creation, FCM push
  - `Pass(ctx, userID, targetID) error` — adds to seen set + swipe rejections
  - `ListMatches(ctx, userID, cursor, limit) ([]*MatchView, string, error)` — pagination clamped to max 20
  - `GetGraphCandidates(ctx, userID) ([]*CandidateCard, error)` — Neo4j PageRank-based candidates
- API endpoints:
  - `GET /v1/matching/candidates` → GetCandidatesHandler
  - `POST /v1/matching/like` → LikeHandler
  - `POST /v1/matching/pass` → PassHandler
  - `GET /v1/matching/likes` → GetPendingLikesHandler
  - `GET /v1/matching/graph-candidates` → GetGraphCandidatesHandler
  - `GET /v1/matches` → ListMatchesHandler
  - `GET /v1/matches/:id` → GetMatchHandler
  - `POST /v1/matches/:id/unmatch` → UnmatchHandler
  - `POST /v1/matches/:id/family-intro` → FamilyIntroHandler
  - `POST /v1/matches/:id/nikah-confirm` → NikahConfirmHandler
- DB tables: social.likes, social.matches, social.swipe_rejections, social.blocks
- Business rules: B1 — niyyah compatibility filter (nikah_year excludes friendship); B2 — same-madhab +10 score boost; B3 — NoPhotoMode candidates have empty AvatarURL + AvatarBlurred=true; bidirectional block exclusion; duplicate match prevention

---

### chat Module
- Handler: `internal/handler/chat_handler.go`
- Service: `internal/service/chat_service.go`
- Repository: `internal/adapter/postgres/message_repo.go`
- Key structs:
  - `Message`: ID uuid, MatchID uuid, SenderID uuid, ContentEncrypted []byte (AES-256-GCM), IsRead bool, IsToxic bool, CreatedAt time.Time
  - `Hub`: clients map[uuid.UUID]*Client, register/unregister chan, broadcast chan, Redis pub/sub subscriber
  - `Client`: conn *websocket.Conn, userID uuid, send chan
- Key service functions:
  - `ListMessages(ctx, matchID, userID, limit, before) ([]*Message, error)` — decrypts content; corrupted rows replaced with "[Хабарлама зақымдалған]" placeholder (logged via slog)
  - `SendMessage(ctx, matchID, senderID, content) (*Message, error)` — AES-256-GCM encrypt, persist, fan-out via Hub
- WebSocket message types (client → server):
  - `auth` — `{"type":"auth","token":"<JWT>"}` (must be first message within 10 seconds)
  - `chat_msg` — `{"type":"chat_msg","match_id":"<uuid>","content":"<text>"}`
  - `mahram_chat_msg` — `{"type":"mahram_chat_msg","room_id":"<uuid>","content":"<text>"}`
  - `typing` — `{"type":"typing","match_id":"<uuid>"}`
  - `read` — `{"type":"read","message_id":"<uuid>"}`
  - `webrtc_offer`, `webrtc_answer`, `webrtc_ice_candidate` — WebRTC signaling
- WebSocket message types (server → client):
  - `auth_ok` — connection confirmed
  - `chat_msg`, `mahram_chat_msg` — incoming messages
  - `content_warning`, `content_blocked` — moderation signals
  - `match_notification` — new match event
  - `typing` — typing indicator
- Hub architecture: goroutine-per-connection; Redis PUBLISH/SUBSCRIBE on channel `chat:<matchID>` for cross-instance routing
- API endpoints:
  - `WS /v1/ws` → WebSocket upgrade → Hub.ServeWS
  - `GET /v1/matches/:matchId/messages` → ListMessagesHandler (cursor pagination)
- DB tables: social.messages

---

### feed Module
- Handler: `internal/handler/post_handler.go`
- Service: `internal/service/post_service.go`
- Repository: `internal/adapter/postgres/post_repo.go`
- Key structs:
  - `Post`: ID uuid, AuthorID uuid, Content string, MediaURL *string, LikeCount int, CommentCount int, CreatedAt time.Time
  - `PostComment`: ID uuid, PostID uuid, AuthorID uuid, Content string, CreatedAt time.Time
- Key service functions:
  - `CreatePost(ctx, authorID, content, mediaURL) (*Post, error)` — reputation gate (score < 30 → 403)
  - `ListPosts(ctx, cursor, limit) ([]*PostWithAuthor, string, error)` — cursor-based pagination
  - `LikePost(ctx, postID, userID) error` / `UnlikePost(...)`
  - `AddComment(ctx, postID, authorID, content) (*PostComment, error)`
  - `ListComments(ctx, postID) ([]*CommentWithAuthor, error)`
- API endpoints:
  - `GET /v1/posts` → ListPostsHandler
  - `POST /v1/posts` → CreatePostHandler (reputation gate: fail-closed on error, score=0)
  - `POST /v1/posts/:id/like` → LikePostHandler
  - `DELETE /v1/posts/:id/like` → UnlikePostHandler
  - `GET /v1/posts/:id/comments` → ListCommentsHandler
  - `POST /v1/posts/:id/comments` → AddCommentHandler
- DB tables: social.posts, social.post_likes, social.comments

---

### settings Module
- Handler: `internal/handler/settings_handler.go`
- Service: `internal/service/settings_service.go`
- Repository: `internal/adapter/postgres/settings_repo.go`
- Key structs:
  - `UserSettings`: UserID string, MinAge int (default 18), MaxAge int (default 60), MaxDistanceKm int (default 50), ShowMe bool, ModestyLevel int (0-3), NiyyahFilter *string, MadhabFilter *string
- Key service functions:
  - `Get(ctx, userID) (*UserSettings, error)` — returns defaults if not set
  - `Upsert(ctx, settings) error`
  - `DefaultSettings(userID) *UserSettings`
- API endpoints:
  - `GET /v1/settings` → GetSettingsHandler
  - `PATCH /v1/settings` → UpdateSettingsHandler (pagination bounds enforced: limit clamped 1–100)
- DB tables: social.user_settings

---

### mahram Module
- Handler: `internal/handler/mahram_handler.go`
- Service: `internal/service/mahram_service.go`
- Repository: `internal/adapter/postgres/mahram_repo.go`
- Key structs:
  - `Mahram`: ID uuid, WomanUserID uuid, MahramPhoneHash []byte, CreatedAt time.Time
  - `MahramRoom`: ID uuid, MatchID uuid, MahramUserID uuid, CreatedAt time.Time
  - `MahramMessage`: ID uuid, RoomID uuid, SenderID uuid, ContentEncrypted []byte (AES-256-GCM), CreatedAt time.Time
- Key service functions:
  - `AddMahram(ctx, womanUserID, phoneHash) (*Mahram, error)` — stores phone hash only, no plaintext
  - `CreateRoom(ctx, matchID, mahramUserID) (*MahramRoom, error)` — validates mahram is third party
  - `ListRooms(ctx, userID) ([]*MahramRoom, error)`
  - `SendMessage(ctx, roomID, senderID, content) (*MahramMessage, error)` — AES-256-GCM encrypt
  - `ListMessages(ctx, roomID, userID) ([]*MahramMessage, error)` — decrypts
- API endpoints:
  - `POST /v1/mahram` → AddMahramHandler
  - `GET /v1/mahram-rooms` → ListRoomsHandler
  - `GET /v1/mahram-rooms/:id/messages` → ListMahramMessagesHandler
- DB tables: social.mahrams, social.mahram_chat_rooms, social.mahram_messages
- Business rules: mahram must be a third party (not one of the match participants); room is identified by matchID + mahramUserID

---

### reputation / interaction Module
- Handler: `internal/handler/interaction_handler.go`
- Service: `internal/service/interaction_service.go`, `internal/service/reputation_service.go`
- Repository: `internal/adapter/postgres/interaction_repo.go`, `internal/adapter/neo4j/trust_graph_repo.go`
- Worker: `internal/worker/trust_engine.go`
- Key structs:
  - `Interaction`: ID uuid, RaterID uuid, RatedID uuid, Context enum (date, meetup, event), Rating int (1-5), Comment *string, ConfirmedAt *time.Time, CreatedAt time.Time
  - `TrustScore`: UserID uuid, Score int (0-100)
  - `LeaderboardEntry`: UserID uuid, DisplayName string, TrustScore int, Rank int
- Trust formula (in Neo4j):
  1. Each rating weighted by rater's KYC level (1.5× if id_verified) × (rater_score / 100)
  2. Bayesian smoothing: score = (Σ weighted_ratings + C × 2.5) / (Σ weights + C) where C=5
  3. Final score = smoothed × 20 → integer 0-100
  4. Written to Neo4j, PostgreSQL (trust_score column), and Redis cache
- Sybil detection: Louvain community detection via Neo4j GDS every 6 hours; clusters with <30% external edges and low KYC rate → flagged in social.sybil_clusters
- API endpoints:
  - `POST /v1/interactions` → SubmitInteractionHandler (rate limiting: 1 rating per user-pair per 24h)
  - `GET /v1/reputation/leaderboard` → LeaderboardHandler (limit clamped 1–100)
- DB tables: social.interactions, social.sybil_clusters; Neo4j: (:User {uid})-[:RATED {rating, weight, created_at}]->(:User)

---

### kyc Module
- Handler: `internal/handler/kyc_handler.go`
- Service: `internal/service/kyc_service.go`
- Repository: `internal/adapter/postgres/kyc_repo.go`, `internal/adapter/minio/media_store.go`
- Key structs:
  - `KYCSubmission`: ID uuid, UserID uuid, Status enum (pending, approved, rejected), DocumentKey string, CreatedAt time.Time
- Key service functions:
  - `Submit(ctx, userID, data []byte, contentType string) error` — MIME validation (image/jpeg, image/png, application/pdf only), MinIO upload, DB record; background goroutine uses server-lifetime context (not context.Background())
  - `GetStatus(ctx, userID) (VerificationLevel, error)`
  - `AdminVerdict(ctx, userID, approved bool) error` — updates verification_level on social.users
- API endpoints:
  - `POST /v1/kyc/submit` → KYCSubmitHandler (multipart form, max 10MB)
  - `GET /v1/kyc/status` → KYCStatusHandler
  - `GET /v1/admin/kyc/pending` → AdminKYCPendingHandler
  - `POST /v1/admin/kyc/:userId/verdict` → AdminKYCVerdictHandler
- DB tables: social.kyc_submissions; identity_vault.iin_vault (restricted PG role, AES-256 encrypted IIN/name)

---

### notifications Module
- Handler: `internal/handler/notification_handler.go`
- Service: `internal/service/notification_service.go`
- Repository: `internal/adapter/postgres/notification_repo.go`
- Key structs:
  - `Notification`: ID uuid, UserID uuid, Type string, Title string, Body string, Data map[string]interface{} (JSONB), IsRead bool, CreatedAt time.Time
- Key service functions:
  - `Create(ctx, userID, notifType, title, body, data) error`
  - `List(ctx, userID, limit) ([]*Notification, error)` — limit clamped 1–100
  - `MarkRead(ctx, notifID, userID) error`
  - `MarkAllRead(ctx, userID) error`
- API endpoints:
  - `GET /v1/notifications` → ListNotificationsHandler
  - `PATCH /v1/notifications/:id/read` → MarkReadHandler
  - `POST /v1/notifications/read-all` → MarkAllReadHandler
- DB tables: social.notifications

---

### whisper Module
- Handler: `internal/handler/whisper_handler.go`
- Service: `internal/service/whisper_service.go`
- Repository: `internal/adapter/postgres/whisper_repo.go`
- Key structs:
  - `WhisperReport`: ID uuid, ReporterID uuid, ReportedID uuid, MatchID *uuid, Content string, Flags int, CreatedAt time.Time
  - `WhisperReportRow`: extends WhisperReport with ID and CreatedAt for admin view
- API endpoints:
  - `POST /v1/whisper` → CreateWhisperHandler (anonymous report, 3-strike system)
  - `GET /v1/admin/whisper-flags` → GetWhisperFlagsHandler (admin only, returns flagged reports)
- DB tables: social.whisper_reports

---

### imam Module
- Handler: `internal/handler/imam_handler.go`
- Package: `internal/pkg/imam/` — embedded catalog
- Implementation: JSON catalog embedded at compile time via `//go:embed` directive; parsed at init-time; zero database round-trips for immutable data
- Catalog: 10 imams across 5 Kazakhstan cities (Astana, Almaty, Shymkent, Karaganda, Aktobe)
- Key functions: `ListByCity(city string) []Imam`, `GetByID(id string) (*Imam, bool)`
- API endpoints:
  - `GET /v1/imams` → ListImamsHandler (optional `?city=` filter)
  - `GET /v1/venues` → ListVenuesHandler (halal venues, optional `?city=` filter)

---

### admin Module
- Handler: `internal/handler/admin_handler.go`
- Service: `internal/service/admin_service.go`
- Middleware: `internal/handler/middleware/auth_middleware.go` (IsAdmin check)
- Key service functions:
  - `ListUsers(ctx, limit, offset) ([]*User, error)`
  - `GetSybilPendingClusters(ctx) ([]*SybilCluster, error)`
  - `GetWhisperFlags(ctx) ([]*WhisperReportRow, error)`
- API endpoints:
  - `GET /v1/admin/users` → AdminListUsersHandler (Admin role)
  - `GET /v1/admin/sybil/pending` → AdminSybilPendingHandler (Admin role)
  - `GET /v1/admin/whisper-flags` → AdminWhisperFlagsHandler (Admin role)

---

### user Module
- Handler: `internal/handler/user_handler.go`
- API endpoints:
  - `GET /v1/users/me` → GetMeHandler
  - `DELETE /v1/users/me` → DeleteMeHandler (soft-delete: is_active=false, cascades matches/posts in transaction)
  - `PATCH /v1/users/me/fcm-token` → UpdateFCMTokenHandler (FCM token for push notifications)
  - `POST /v1/users/:id/block` → BlockUserHandler (writes to social.blocks + calls UnmatchByUsers)

---

## Database Schema

### social.users
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY DEFAULT gen_random_uuid() |
| phone_hash | bytea | UNIQUE NOT NULL |
| phone_encrypted | bytea | NOT NULL |
| email_encrypted | bytea | |
| password_hash | text | NOT NULL |
| public_key | text | |
| verification_level | text | DEFAULT 'none' CHECK IN (none, phone_verified, id_verified, photo_verified) |
| trust_status | text | DEFAULT 'normal' CHECK IN (normal, under_review, suspended, banned) |
| trust_score | int | DEFAULT 0 CHECK (0-100) |
| is_admin | bool | DEFAULT false |
| is_active | bool | DEFAULT true |
| last_login_at | timestamptz | |
| fcm_token | text | |
| created_at | timestamptz | DEFAULT NOW() |
| updated_at | timestamptz | DEFAULT NOW() |

### social.profiles
| Column | Type | Constraints |
|--------|------|-------------|
| user_id | uuid | PRIMARY KEY REFERENCES social.users(id) ON DELETE CASCADE |
| display_name | text | |
| bio | text | |
| gender | text | CHECK IN (male, female, other) |
| birth_date | date | |
| city | text | |
| location | geometry(Point, 4326) | PostGIS |
| looking_for | text | CHECK IN (male, female, other) |
| avatar_url | text | |
| niyyah | text | CHECK IN (nikah_year, serious_marriage, friendship) |
| madhab | text | CHECK IN (hanafi, shafii, maliki, hanbali, none) |
| languages | text[] | DEFAULT '{}' |
| no_photo_mode | bool | DEFAULT false |
| marital_status | text | DEFAULT 'single' |
| prompts | jsonb | DEFAULT '[]' |
| created_at | timestamptz | DEFAULT NOW() |
| updated_at | timestamptz | DEFAULT NOW() |

### social.likes
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| liker_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| liked_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| created_at | timestamptz | DEFAULT NOW() |
Foreign Keys: UNIQUE(liker_id, liked_id)

### social.matches
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| user_a_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| user_b_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| matched_at | timestamptz | DEFAULT NOW() |
| niyyah_timer_ends_at | timestamptz | |
| family_intro_done | bool | DEFAULT false |
| imam_confirmed | bool | DEFAULT false |
Foreign Keys: UNIQUE(user_a_id, user_b_id)

### social.messages
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| match_id | uuid | REFERENCES social.matches(id) ON DELETE CASCADE |
| sender_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| content | bytea | NOT NULL (AES-256-GCM encrypted) |
| is_read | bool | DEFAULT false |
| is_toxic | bool | DEFAULT false |
| created_at | timestamptz | DEFAULT NOW() |
Indexes: idx_messages_match (match_id, created_at DESC), idx_messages_sender (sender_id, created_at DESC), idx_messages_toxic (is_toxic, created_at DESC) WHERE is_toxic = true

### social.posts
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| author_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| content | text | NOT NULL |
| media_url | text | |
| like_count | int | DEFAULT 0 |
| comment_count | int | DEFAULT 0 |
| created_at | timestamptz | DEFAULT NOW() |

### social.post_likes
| Column | Type | Constraints |
|--------|------|-------------|
| post_id | uuid | REFERENCES social.posts(id) ON DELETE CASCADE |
| user_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| created_at | timestamptz | DEFAULT NOW() |
Primary Key: (post_id, user_id)

### social.comments
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| post_id | uuid | REFERENCES social.posts(id) ON DELETE CASCADE |
| author_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| content | text | NOT NULL |
| created_at | timestamptz | DEFAULT NOW() |

### social.user_settings
| Column | Type | Constraints |
|--------|------|-------------|
| user_id | uuid | PRIMARY KEY REFERENCES social.users(id) ON DELETE CASCADE |
| min_age | int | DEFAULT 18 |
| max_age | int | DEFAULT 60 |
| max_distance_km | int | DEFAULT 50 |
| show_me | bool | DEFAULT true |
| modesty_level | int | DEFAULT 0 CHECK (0-3) |
| niyyah_filter | text | |
| madhab_filter | text | |

### social.mahrams
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| woman_user_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| mahram_phone_hash | bytea | NOT NULL |
| created_at | timestamptz | DEFAULT NOW() |

### social.mahram_chat_rooms
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| match_id | uuid | REFERENCES social.matches(id) ON DELETE CASCADE |
| mahram_user_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| created_at | timestamptz | DEFAULT NOW() |

### social.mahram_messages
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| room_id | uuid | REFERENCES social.mahram_chat_rooms(id) ON DELETE CASCADE |
| sender_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| content | bytea | NOT NULL (AES-256-GCM encrypted) |
| created_at | timestamptz | DEFAULT NOW() |

### social.swipe_rejections
| Column | Type | Constraints |
|--------|------|-------------|
| user_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| rejected_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| created_at | timestamptz | DEFAULT NOW() |
Primary Key: (user_id, rejected_id)

### social.blocks
| Column | Type | Constraints |
|--------|------|-------------|
| blocker_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| blocked_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| created_at | timestamptz | DEFAULT NOW() |
Primary Key: (blocker_id, blocked_id)

### social.notifications
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| user_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| type | text | NOT NULL |
| title | text | NOT NULL |
| body | text | NOT NULL |
| data | jsonb | DEFAULT '{}' |
| is_read | bool | DEFAULT false |
| created_at | timestamptz | DEFAULT NOW() |
Indexes: idx_notifications_user (user_id, is_read, created_at DESC)

### social.whisper_reports
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| reporter_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| reported_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| match_id | uuid | REFERENCES social.matches(id) ON DELETE SET NULL |
| content | text | |
| flags | int | DEFAULT 0 |
| created_at | timestamptz | DEFAULT NOW() |

### social.kyc_submissions
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| user_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| status | text | DEFAULT 'pending' CHECK IN (pending, approved, rejected) |
| document_key | text | (MinIO object key) |
| created_at | timestamptz | DEFAULT NOW() |

### social.refresh_tokens
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| user_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| token_hash | bytea | UNIQUE NOT NULL (Argon2id hashed) |
| expires_at | timestamptz | NOT NULL |
| used_at | timestamptz | |
Indexes: idx_refresh_tokens_user (user_id, expires_at)

### social.interactions
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| rater_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| rated_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| context | text | CHECK IN (date, meetup, event) |
| rating | int | CHECK (1-5) |
| comment | text | |
| confirmed_at | timestamptz | |
| created_at | timestamptz | DEFAULT NOW() |
Indexes: idx_interactions_rater (rater_id, created_at DESC), idx_interactions_rated (rated_id, created_at DESC)
Rate limiting: 1 rating per (rater_id, rated_id) per 24 hours (Redis cooldown key)

### social.sybil_clusters
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| cluster_id | text | |
| user_ids | uuid[] | |
| flagged | bool | DEFAULT false |
| created_at | timestamptz | DEFAULT NOW() |

### social.media
| Column | Type | Constraints |
|--------|------|-------------|
| id | uuid | PRIMARY KEY |
| user_id | uuid | REFERENCES social.users(id) ON DELETE CASCADE |
| object_key | text | NOT NULL (MinIO object key) |
| content_type | text | |
| created_at | timestamptz | DEFAULT NOW() |

### identity_vault.iin_vault (restricted PG role)
| Column | Type | Constraints |
|--------|------|-------------|
| user_id | uuid | PRIMARY KEY REFERENCES social.users(id) |
| iin_encrypted | bytea | AES-256-GCM |
| iin_hash | bytea | UNIQUE (for uniqueness check without decryption) |
| full_name_encrypted | bytea | AES-256-GCM |
| created_at | timestamptz | DEFAULT NOW() |

---

## API Endpoints — Complete List

| Method | Path | Auth Required | Role | Description |
|--------|------|---------------|------|-------------|
| POST | /v1/auth/register | No | — | Register with phone+password |
| POST | /v1/auth/login | No | — | Login, receive JWT + refresh cookie |
| POST | /v1/auth/refresh | Cookie | — | Rotate refresh token |
| POST | /v1/auth/logout | Yes | — | Invalidate refresh token |
| GET | /v1/users/me | Yes | — | Get current user info |
| DELETE | /v1/users/me | Yes | — | Soft-delete account |
| PATCH | /v1/users/me/fcm-token | Yes | — | Register FCM push token |
| POST | /v1/users/:id/block | Yes | — | Block a user (removes match) |
| GET | /v1/profiles/me | Yes | — | Get own profile |
| PATCH | /v1/profiles/me | Yes | — | Update own profile |
| GET | /v1/profiles/:userId | Yes | — | Get another user's profile |
| POST | /v1/profiles/me/photos | Yes | — | Upload profile photo to MinIO |
| GET | /v1/matching/candidates | Yes | — | Get discovery candidates (PostGIS + filters) |
| POST | /v1/matching/like | Yes | — | Like a candidate; returns matched+matchId |
| POST | /v1/matching/pass | Yes | — | Pass on a candidate |
| GET | /v1/matching/likes | Yes | — | Get pending likes (users who liked you) |
| GET | /v1/matching/graph-candidates | Yes | — | Get Neo4j PageRank-based candidates |
| GET | /v1/matches | Yes | — | List all matches (cursor pagination) |
| GET | /v1/matches/:id | Yes | — | Get single match details |
| POST | /v1/matches/:id/unmatch | Yes | — | Unmatch and remove conversation |
| POST | /v1/matches/:id/family-intro | Yes | — | Mark family introduction done |
| POST | /v1/matches/:id/nikah-confirm | Yes | — | Confirm nikah with imam |
| GET | /v1/matches/:matchId/messages | Yes | — | Get chat message history |
| WS | /v1/ws | JWT in first msg | — | WebSocket: real-time chat + signals |
| POST | /v1/interactions | Yes | — | Submit post-meeting interaction rating |
| GET | /v1/reputation/leaderboard | Yes | — | Top trust score leaderboard |
| GET | /v1/posts | Yes | — | Get community feed (cursor paginated) |
| POST | /v1/posts | Yes | — | Create post (rep gate: score ≥ 30) |
| POST | /v1/posts/:id/like | Yes | — | Like a post |
| DELETE | /v1/posts/:id/like | Yes | — | Unlike a post |
| GET | /v1/posts/:id/comments | Yes | — | Get post comments |
| POST | /v1/posts/:id/comments | Yes | — | Add comment to post |
| POST | /v1/kyc/submit | Yes | — | Submit ID documents (multipart) |
| GET | /v1/kyc/status | Yes | — | Get own KYC verification level |
| GET | /v1/notifications | Yes | — | List notifications (limit ≤ 100) |
| PATCH | /v1/notifications/:id/read | Yes | — | Mark notification as read |
| POST | /v1/notifications/read-all | Yes | — | Mark all notifications as read |
| GET | /v1/settings | Yes | — | Get discovery preferences |
| PATCH | /v1/settings | Yes | — | Update discovery preferences |
| POST | /v1/mahram | Yes | — | Add a mahram (guardian) by phone hash |
| GET | /v1/mahram-rooms | Yes | — | List mahram chat rooms |
| GET | /v1/mahram-rooms/:id/messages | Yes | — | Get mahram chat history |
| POST | /v1/whisper | Yes | — | Submit anonymous content report |
| GET | /v1/imams | Yes | — | List imams (optionally by city) |
| GET | /v1/venues | Yes | — | List halal venues (optionally by city) |
| GET | /v1/admin/users | Yes | Admin | List all users |
| GET | /v1/admin/kyc/pending | Yes | Admin | List pending KYC submissions |
| POST | /v1/admin/kyc/:userId/verdict | Yes | Admin | Approve/reject KYC |
| GET | /v1/admin/sybil/pending | Yes | Admin | List flagged Sybil clusters |
| GET | /v1/admin/whisper-flags | Yes | Admin | List flagged whisper reports |

---

## WebSocket Protocol

- Connection URL: `ws://<host>/v1/ws` (production: `wss://`)
- Auth mechanism: After TCP+TLS upgrade, client must send `{"type":"auth","token":"<JWT access token>"}` as first message within 10 seconds; backend closes connection if absent or invalid
- Why first-message auth (not URL param): URL query parameters appear in server logs; embedding token in first encrypted message is more secure

**Message types (client → server):**
| Type | JSON Structure | Purpose |
|------|---------------|---------|
| auth | `{"type":"auth","token":"<jwt>"}` | Authentication handshake |
| chat_msg | `{"type":"chat_msg","match_id":"<uuid>","content":"<text>"}` | Send chat message |
| mahram_chat_msg | `{"type":"mahram_chat_msg","room_id":"<uuid>","content":"<text>"}` | Send mahram channel message |
| typing | `{"type":"typing","match_id":"<uuid>"}` | Typing indicator |
| read | `{"type":"read","message_id":"<uuid>"}` | Mark message read |
| webrtc_offer | `{"type":"webrtc_offer","match_id":"<uuid>","sdp":"<sdp>"}` | WebRTC call offer |
| webrtc_answer | `{"type":"webrtc_answer","match_id":"<uuid>","sdp":"<sdp>"}` | WebRTC call answer |
| webrtc_ice_candidate | `{"type":"webrtc_ice_candidate","match_id":"<uuid>","candidate":"<cand>","sdpMid":"<mid>","sdpMLineIndex":<n>}` | ICE candidate |

**Message types (server → client):**
| Type | JSON Structure | Purpose |
|------|---------------|---------|
| auth_ok | `{"type":"auth_ok"}` | Connection confirmed |
| chat_msg | `{"type":"chat_msg","id":"<uuid>","match_id":"<uuid>","sender_id":"<uuid>","content":"<text>","is_read":false,"created_at":"<RFC3339>"}` | Incoming message |
| mahram_chat_msg | `{"type":"mahram_chat_msg","id":"<uuid>","room_id":"<uuid>","sender_id":"<uuid>","content":"<text>","created_at":"<RFC3339>"}` | Mahram channel message |
| content_warning | `{"type":"content_warning","match_id":"<uuid>","id":"<uuid>","content":"<text>","sender_id":"<uuid>","created_at":"<RFC3339>"}` | Flagged content |
| content_blocked | `{"type":"content_blocked","match_id":"<uuid>"}` | Message blocked by policy |
| match_notification | `{"type":"match_notification","match_id":"<uuid>","other_user_id":"<uuid>"}` | New mutual match |
| typing | `{"type":"typing","match_id":"<uuid>","sender_id":"<uuid>"}` | Typing indicator relay |

**Hub structure:**
- `clients: map[uuid.UUID]*Client` — one entry per connected user
- `register chan *Client` — goroutine sends new client on connect
- `unregister chan *Client` — goroutine sends on disconnect
- `broadcast chan []byte` — internal message dispatch
- Each Client runs `readPump` and `writePump` goroutines
- Panic recovery in readLoop prevents Hub crash from one bad connection

**Redis pub/sub:**
- Channel: `chat:<matchID>` (string)
- On outbound message: Hub PUBLISH to Redis channel
- Background goroutine SUBSCRIBE to all channels; RECEIVE → forward to local Hub
- Enables horizontal scaling across multiple API instances

---

## Frontend Screens

| Screen File | Route | Purpose | Provider Used |
|------------|-------|---------|--------------|
| splash_screen.dart | /splash | Animated Islamic star (800ms), auth redirect | authStateProvider |
| onboarding_screen.dart | /onboarding | 3-slide intro with CustomPainter illustrations | — |
| niyyah_selection_screen.dart | /niyyah | Select marriage intention (3 cards, dark bg, gold border) | — |
| login_screen.dart | /auth/login | Phone + password login, Kazakh UI | authStateProvider |
| register_screen.dart | /auth/register | Signup: phone, password, gender, madhab | authStateProvider |
| home_shell.dart | (ShellRoute) | GoRouter ShellRoute, custom 5-tab bottom navigation | — |
| discovery_screen.dart | /home | FlutterCardSwiper candidates, filter bottom sheet, match banner | matchingNotifierProvider |
| feed_screen.dart | /feed | Social post feed with like/comment | feedProvider |
| create_post_screen.dart | /feed/create | Compose and publish new post | feedProvider |
| matches_screen.dart | /matches | Match list with niyyah timer countdown chips (orange/red) | matchesListProvider |
| profile_screen.dart | /profile | Own profile: avatar, trust badge, niyyah/madhab/languages, KYC status, completeness bar, logout | ownProfileProvider |
| edit_profile_screen.dart | /profile/edit | Edit all profile fields, birth_date picker, language selector | ownProfileProvider |
| profile_detail_screen.dart | /profile/:userId | View other user's profile, whisper ghost button | — |
| settings_screen.dart | /settings | Age range, distance, modesty, niyyah/madhab filters, mahram management | settingsNotifierProvider |
| notifications_screen.dart | /notifications | Notification list with read/unread state | notificationsProvider |
| chat_screen.dart | /chat/:matchId | Real-time WebSocket chat, read receipts, typing indicator, mahram invite bottom sheet | chatNotifierProvider |
| mahram_chat_screen.dart | /mahram-chat/:roomId | 3-way chat: woman=green/man=blue/mahram=gold bubbles | chatNotifierProvider |
| call_screen.dart | /call/:matchId | WebRTC video/audio call (flutter_webrtc) | — |
| kyc_screen.dart | /kyc | ID photo upload via image_picker, multipart POST | kycStatusProvider |
| imam_connect_screen.dart | /imams | City dropdown, imam cards, nikah confirm dialog | — |
| first_meeting_screen.dart | /first-meeting/:matchId | Venue selection + interaction rating | matchingNotifierProvider |

---

## Frontend Providers (State Management)

| Provider | Type | State Type | Key Methods |
|----------|------|-----------|------------|
| authStateProvider | StateNotifierProvider | AsyncValue\<User?\> | login(), register(), logout(), _restoreSession() |
| dioClientProvider | Provider | DioClient | — (configured with interceptors) |
| matchingNotifierProvider | StateNotifierProvider | AsyncValue\<List\<Map\>\> | load(), like(id), pass(id) |
| pendingLikesProvider | FutureProvider | List\<Map\> | — (re-fetches on auth change) |
| matchesListProvider | FutureProvider | List\<Map\> | — |
| feedProvider | StateNotifierProvider | AsyncValue\<List\<Post\>\> | load(), loadMore(), likePost(), createPost() |
| postCommentsProvider | FutureProvider.family | List\<PostComment\> | keyed by postId |
| chatNotifierProvider | StateNotifierProvider.family | ChatState | send(), sendTyping(), sendReadReceipt() |
| settingsNotifierProvider | StateNotifierProvider | AsyncValue\<UserSettings\> | update() — debounced 450ms, reverts state + shows snackbar on error |
| notificationsProvider | StateNotifierProvider | AsyncValue\<List\<AppNotification\>\> | load(), markRead(id), markAllRead() |
| unreadCountProvider | FutureProvider | int | — |
| ownProfileProvider | FutureProvider | Profile | — |
| kycStatusProvider | FutureProvider | String | — (returns 'none' or 'id_verified') |

**DioClient interceptors (in order):**
1. `_AuthInterceptor` — adds `Authorization: Bearer <token>`; on 401 → POST /auth/refresh → retry; on refresh failure → deleteAll + logout
2. `_RetryInterceptor` — max 3 retries, exponential backoff 100ms × 2^attempt; triggers on connection errors, timeouts, 5xx
3. `_ErrorInterceptor` — shows snackbar for network/5xx errors; skips 401

**ChatNotifier reconnection:** exponential backoff 1s/2s/4s/8s/16s, max 5 attempts before stopping

---

## Key Model Classes (Frontend)

```dart
User { id, phone, email, isKycVerified, isAdmin, createdAt }

Profile {
  userId, displayName, age, city, bio,
  avatarUrl, avatarBlurred, trustScore, isKycVerified,
  gender, lookingFor, niyyah, madhab, languages[],
  prompts[{question, answer}], noPhotoMode, maritalStatus
}

Match {
  id, otherUserId, otherUserName, otherUserAvatarUrl,
  otherUserTrustScore, status,
  niyyahTimerEndsAt, familyIntroDone, imamConfirmed, createdAt,
  daysLeftOnTimer (computed: max(0, diff))  // never negative
}

Post {
  id, authorId, content, mediaUrl,
  likeCount, commentCount, createdAt,
  authorName, authorAvatarUrl, isLiked
}

PostComment { id, postId, authorId, content, createdAt, authorName, authorAvatarUrl }

UserSettings {
  minAge(18), maxAge(60), maxDistanceKm(50),
  showMe(bool), modestyLevel(0-3),
  niyyahFilter(nullable), madhabFilter(nullable)
}

AppNotification { id, userId, type, title, body, data, isRead, createdAt }

Interaction { id, fromUserId, toUserId, context(date|meetup|event), rating(1-5), comment, confirmedAt }

Venue { id, name, city, category(cafe|restaurant|park|cultural), address, phone }

AuthTokens { accessToken, refreshToken, user }
```

**Token storage:** `flutter_secure_storage` keys — `access_token`, `refresh_token`, `cached_user` (JSON)

**API base URL:** `http://localhost:8080/v1` (dev); configurable by platform  
**WebSocket URL:** `ws://localhost:8080/v1/ws`

---

## Infrastructure

### Docker Services (deployments/docker-compose.yml)

| Service | Image | Ports | Key Config |
|---------|-------|-------|------------|
| migrate | Built from Dockerfile | — | Runs `/bin/migrate up`; depends_on: postgres(healthy) |
| api | Built from Dockerfile | 8080:8080 | Depends on: migrate, postgres, redis, neo4j, minio |
| worker | Built from Dockerfile | — | Entrypoint: `/bin/worker` |
| postgres | postgis/postgis:16-3.4-alpine | 5433:5432 | Volume: pg_data; healthcheck: pg_isready |
| neo4j | neo4j:5-community | 7474:7474, 7687:7687 | Plugins: GDS; volumes: neo4j_data, neo4j_plugins; healthcheck: wget 7474 |
| redis | redis:7-alpine | 6379:6379 | `--maxmemory 256mb --maxmemory-policy allkeys-lru`; healthcheck: redis-cli ping |
| minio | minio/minio:latest | 9000:9000, 9001:9001 | Volume: minio_data; healthcheck (production): curl /minio/health/live |
| nginx | nginx:alpine | 80:80 | Mounts nginx.conf; depends_on: api |

### Dockerfile (deployments/Dockerfile)
**Stage 1 — Build (golang:1.25-alpine):**
- Installs `git ca-certificates`
- COPY go.mod/go.sum, then full source
- Builds `/bin/api`, `/bin/worker`, `/bin/migrate` with `CGO_ENABLED=0 GOOS=linux -ldflags="-s -w"`

**Stage 2 — Runtime (alpine:3.19):**
- Installs `ca-certificates tzdata`
- Creates non-root user `appuser` (UID 1000)
- COPY binaries + migrations from builder stage
- EXPOSE 8080, ENTRYPOINT `/bin/api`

### nginx Configuration (deployments/nginx/nginx.conf)
**Upstream:** `api_backend { server api:8080; }`

**Rate limiting zones:**
- `api_limit`: 30 req/s (10MB zone)
- `auth_limit`: 5 req/s (10MB zone) — applied to `/v1/auth/`

**Location blocks:**
- `/v1/` — rate limit api_limit burst=20; proxy to api_backend
- `/v1/auth/` — rate limit auth_limit burst=5 (stricter for brute-force protection)
- `/v1/ws` — WebSocket upgrade (`Upgrade: $http_upgrade`, `Connection: upgrade`); proxy_read/send_timeout 86400s
- `/v1/health` — no rate limit

**Security headers:** X-Content-Type-Options: nosniff, X-Frame-Options: DENY, X-XSS-Protection: 1; mode=block, Referrer-Policy: strict-origin-when-cross-origin

**Max request body:** 10m (for photo/document uploads)

### CI/CD Pipeline (.github/workflows/ci.yml)
**Trigger:** push to main, pull_request to main  
**Jobs (ubuntu-latest):**
1. Checkout (`actions/checkout@v4`)
2. Set up Go 1.22 (`actions/setup-go@v5`)
3. `go build ./cmd/api`
4. `go build ./cmd/worker`
5. `go test ./internal/... -v -race -count=1`
6. `go vet ./...`

### Makefile Targets
| Target | Command |
|--------|---------|
| build | Build api, worker, migrate binaries to bin/ |
| run | `go run ./cmd/api` |
| run-worker | `go run ./cmd/worker` |
| test | `go test ./... -v -race -count=1` |
| lint | `golangci-lint run ./...` |
| migrate-up | `go run ./cmd/migrate up` |
| migrate-down | `go run ./cmd/migrate down` |
| docker-up | `docker compose -f deployments/docker-compose.yml --env-file .env up --build -d` |
| docker-down | `docker compose down` |
| docker-logs | `docker compose logs -f` |
| docker-reset | Down -v + up --build (wipes all data) |
| seed-halal | Load demo users from migrations/seed_halal_demo.sql |
| reset-demo | Full reset + seed with 12s wait |

### Environment Variables (.env.example)
```
APP_ENV, SERVER_HOST, SERVER_PORT, CORS_ORIGINS
PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DBNAME, PG_SSLMODE
NEO4J_URI, NEO4J_USER, NEO4J_PASSWORD
REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB
MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_USE_SSL, MINIO_BUCKET
JWT_SECRET (min 32 chars), JWT_ACCESS_EXPIRY, JWT_REFRESH_EXPIRY
ENCRYPTION_KEY (32-byte AES key, hex-encoded; generate: openssl rand -hex 32)
FIREBASE_CREDENTIALS (path to Firebase service account JSON)
```

---

## Islamic Domain Features

### Niyyah (نية — Marriage Intention)
- Values: `nikah_year` (planning marriage within ~1 year), `serious_marriage` (marriage-focused), `friendship` (casual/friendship)
- Stored: `social.profiles.niyyah` (text column with CHECK constraint)
- Filter (B1): Discovery query applies niyyah compatibility — nikah_year and serious_marriage users do not see friendship candidates; configured via `AllowedNiyyahs` list in `FindCandidatesOpts`
- Settings: `social.user_settings.niyyah_filter` — user can restrict their pool further
- UI: `NiyyahBadge` widget (chip with color coding), `NiyyahSelectionScreen` during onboarding
- Timer: `niyyah_timer_ends_at` on matches — 90-day expiry for nikah_year matches; `NiyyahTimerWorker` checks daily; `daysLeftOnTimer` computed property (min 0, never negative); UI shows orange (>3 days) or red (≤3 days) countdown chip

### Madhab (مذهب — Islamic School of Jurisprudence)
- Values: `hanafi`, `shafii`, `maliki`, `hanbali`, `none`
- Stored: `social.profiles.madhab`
- Boost (B2): Same-madhab candidates receive +10 score in discovery ranking
- Settings: `social.user_settings.madhab_filter`
- UI: `MadhabBadge` widget; shown on profile cards and profile detail screen

### NoPhotoMode (حشمة — Modesty)
- Feature (B3): Users can enable `no_photo_mode` on their profile
- Behavior: Discovery query returns `AvatarURL = ""` and `NoPhotoMode = true` for these users
- Flutter: `CandidateCard` renders blurred/placeholder image when `avatarBlurred = true`
- Purpose: Allows women to control photo visibility — first interaction based on personality/values

### Mahram (محرم — Guardian Supervision)
- Context: In Islamic law, unmarried men and women should not communicate privately without a mahram (close male relative or legal guardian) present
- Implementation: `social.mahrams` stores woman's user_id + guardian's phone hash (not plaintext); `social.mahram_chat_rooms` creates a 3-way chat channel (woman + man + mahram); messages AES-256-GCM encrypted
- WS integration: `mahram_chat_msg` type routes through Hub to all three participants
- UI: `MahramChatScreen` (green bubbles for woman, blue for man, gold for mahram guardian); mahram invite bottom sheet in ChatScreen
- Access control: Room creation validates mahram is a third party (not a match participant)

### Family Intro and Imam Confirmation
- `family_intro_done`: Boolean flag on `social.matches`; set via `POST /v1/matches/:id/family-intro`; represents formal family introduction step in Islamic courtship
- `imam_confirmed`: Boolean flag set via `POST /v1/matches/:id/nikah-confirm`; records that an imam witnessed the nikah agreement; also sets `married_via_app = true` on both users' profiles
- Imam catalog: Embedded at compile time (`internal/pkg/imam/`), 10 imams across 5 KZ cities; `ListByCity()` and `GetByID()` functions; no database required

### KYC and Identity Vault
- Documents: Image uploads (JPEG, PNG, PDF) to MinIO; max 10MB
- Identity vault: Separate PostgreSQL schema (`identity_vault`) with restricted role; IIN (Individual Identification Number) encrypted with AES-256-GCM; IIN hash stored for uniqueness checks without decryption
- Trust weight: KYC-verified users get 1.5× weight multiplier in trust score formula
- Verification levels: none → phone_verified → id_verified → photo_verified

### Trust Score Engine
- Range: 0-100 integer
- Formula: Bayesian smoothing → (Σ(rating_i × weight_i) + C × 2.5) / (Σweight_i + C) × 20, where C=5 virtual ratings pull toward neutral 2.5/5
- Weight per rating: `(rater_kyc_multiplier) × (rater_trust_score / 100)` — KYC-verified rater gets 1.5× boost
- Storage: Written to Neo4j graph (:User {uid})-[:RATED]->(:User), PostgreSQL (social.users.trust_score), Redis cache
- Computation: Async in `TrustEngine` worker via buffered channel (`chan uuid.UUID` capacity 100); drains remaining events on shutdown
- Sybil detection: Louvain community detection (Neo4j GDS) runs every 6 hours; clusters with <30% external edges + low KYC rate → `social.sybil_clusters.flagged = true`

### Whisper Reports (Anonymous Reporting)
- Purpose: Allow users to anonymously report inappropriate behavior (3-strike philosophy — pattern detection, not single report)
- Storage: `social.whisper_reports` with `flags` counter (increments on each report against same user)
- Admin review: `GET /v1/admin/whisper-flags` returns reports with flag count; admin can escalate to account suspension

---

## Known Code Issues / Technical Debt (resolved in audit — all 102+ tests pass)

- B4: Neo4j `DeleteUserNode` now uses `{uid: $uid}` (was incorrectly `{id: $uid}`)
- B12: PushWorker nil FCMToken panic fixed — `if user.FCMToken == nil { continue }`
- B2: `BlockUser()` now calls `UnmatchByUsers()` to remove the match card
- B7: Discovery filter Apply button now calls `settingsNotifier.update()` and invalidates candidate cache
- B8: WebSocket ChatNotifier implements exponential backoff reconnect (1s/2s/4s/8s/16s, max 5 attempts)
- B6: Corrupted encrypted messages replaced with "[Хабарлама зақымдалған]" placeholder (slog.Warn logged)
- D2: Migration 000019 adds ON DELETE CASCADE to social.interactions, social.messages, social.whisper_reports
- S4: Pagination `limit` parameter clamped to 1–100 in all three handlers (matching, interaction, admin)
- L7: Read receipts auto-fired via `ref.listen` on chatNotifierProvider for last visible message from other user
- L9: Settings update reverts to previous state + shows snackbar on network error
- L4: Chat message deduplication uses `indexWhere` (not `lastIndexWhere`) to confirm oldest pending temp message first
- F1: Push notifications sent for both likes and mutual matches via FCM push channel
- F2: Profile completeness bar widget (`_ProfileCompletenessBar`) on ProfileScreen (checks 8 fields, 3-color scale)
- Migration 000018: Performance indexes on cooldown and trust-score columns
