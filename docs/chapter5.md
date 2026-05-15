# Chapter 5: System Architecture and Design

## 5.1 System Architecture Overview

TrueConnect is built on a three-tier architecture in which each tier has a clearly defined responsibility, a well-specified communication protocol, and a technology stack chosen to match its performance requirements.

**Tier 1 — Flutter Mobile Client** handles all user interaction, presentation logic, and client-side state management. The Flutter application communicates with the backend exclusively through two channels: HTTPS REST API calls (all CRUD operations, profile management, feed, notifications, settings) and a persistent WebSocket connection over WSS (real-time chat, typing indicators, read receipts, match notifications, and WebRTC signaling). Authentication tokens are stored in `flutter_secure_storage` and injected automatically by the Dio HTTP client's `_AuthInterceptor`. State management is handled by Riverpod providers, which cache API responses in memory and propagate changes to the widget tree through Dart streams.

**Tier 2 — Go Backend (Modular Monolith)** contains all business logic, data access, and real-time message routing. The backend is a single deployable binary (`/bin/api`) that exposes both a Gin HTTP server and a WebSocket Hub on port 8080. nginx sits in front of the API as a reverse proxy, handling TLS termination, rate limiting, WebSocket upgrade header injection, and request routing. A separate worker binary (`/bin/worker`) runs background goroutines for trust score computation, Sybil detection, FCM push delivery, niyyah timer expiry, and refresh token cleanup.

**Tier 3 — Data Layer** consists of four specialised stores: PostgreSQL 16 with PostGIS for relational data and geospatial queries; Neo4j 5 Community for the trust graph and Sybil detection algorithms; Redis 7 for WebSocket pub/sub fan-out, rate limiting counters, matching candidate seen-sets, and trust score caching; and MinIO for object storage of profile photos and KYC documents.

The architecture was chosen as a modular monolith rather than microservices because it provides clear module boundaries (15 Go packages with interface-based communication) without the distributed systems overhead of service discovery, network partitions, and distributed tracing that would be disproportionate for the current scale and team size.

---

## 5.2 Backend Clean Architecture Layers

The backend implements five layers as described by Robert C. Martin's Clean Architecture, with the dependency rule strictly enforced: inner layers are never aware of outer layers.

**Domain Layer** (`internal/domain/`): The innermost layer. Contains pure Go structs and typed constants with zero external imports. Key types include `domain.User` (identity entity), `domain.Profile` (Islamic profile entity with `Niyyah Niyyah` and `Madhab Madhab` typed fields), `domain.Match`, `domain.Message`, `domain.Post`, `domain.Interaction`, `domain.Mahram`, `domain.MahramRoom`. All error sentinels are defined here: `domain.ErrNotFound`, `domain.ErrForbidden`, `domain.ErrInvalidInput`, `domain.ErrConflict`. Business invariants (e.g., niyyah enum values, madhab enum values) are expressed as Go type constraints at this layer and can never be violated by any outer layer.

**Repository Layer** (`internal/repository/`): Interface definitions (ports) that the service layer depends on. Each interface corresponds to a domain aggregate. Examples: `repository.ProfileRepository` with `Upsert(ctx, *domain.Profile) error` and `FindCandidates(ctx, FindCandidatesOpts) ([]*CandidateRow, error)`; `repository.MatchRepository` with `RecordLike(ctx, userID, targetID uuid.UUID) (bool, uuid.UUID, error)`. The `FindCandidatesOpts` struct is also defined in this layer: `RequesterID uuid.UUID`, `ExcludeIDs []uuid.UUID`, `AllowedNiyyahs []string`, `Limit int` — making the niyyah filter a typed contract between the service and repository layers.

**Service Layer** (`internal/service/`): Business logic orchestration. Services receive repository interfaces via constructor injection and never import adapter packages. `matching_service.go` demonstrates the pattern: `GetCandidates()` calls `settingsRepo.Get()` to load the requester's settings, constructs `FindCandidatesOpts` with the niyyah compatibility list, calls `profileRepo.FindCandidates()`, then applies the madhab affinity boost (+10 to `TrustScore` for same-madhab candidates) and the no-photo blur (sets `AvatarURL = ""` and `AvatarBlurred = true` when `CandidateRow.NoPhotoMode == true`) before returning `CandidateView` structs to the handler. The `candidateBatchSize` constant (20) and `seenSetTTL` (24 hours) are defined in this layer.

**Handler Layer** (`internal/handler/`): Gin HTTP handlers, middleware, and the WebSocket Hub. Handlers extract validated request parameters, call the appropriate service method, and write JSON responses using a consistent envelope format: `{"status":"success","data":{...}}` or `{"status":"error","message":"..."}`. Auth middleware (`middleware/auth_middleware.go`) extracts the JWT Bearer token, verifies it using `tcjwt.Manager`, and injects the `userID` into the Gin context. The `Hub` struct in `chat_handler.go` manages WebSocket connections: `connections map[uuid.UUID]*websocket.Conn`, protected by `sync.RWMutex`, with injected references to `*service.ChatService`, `*service.MatchingService`, `*service.ReputationService`, `*service.MahramChatService`, a `*redis.Client` for pub/sub, and a `*tcjwt.Manager` for token verification.

**Adapter Layer** (`internal/adapter/`): Concrete implementations of repository interfaces. `adapter/postgres/` contains all SQL query implementations using `pgx/v5`. `adapter/neo4j/` implements the trust graph operations (Cypher queries). `adapter/redis/` implements cache and pub/sub operations. `adapter/minio/` implements the `MediaStore` interface for object storage. Adapters are the only layer that imports external database drivers; they are never imported by services.

---

## 5.3 Backend Module Architecture

### Auth Module
`AuthService.Register()` validates the phone number format, hashes the password with Argon2id (64MB memory, 3 iterations, 4 parallelism threads), encrypts the phone and email fields with AES-256-GCM (12-byte random nonce per field, stored prepended to ciphertext), creates the user record in `social.users`, issues a JWT access token (HS256, 15-minute expiry, claims: `user_id`, `is_admin`) and a refresh token (256-bit random bytes, Argon2id-hashed before storage in `social.refresh_tokens`). `AuthService.Login()` retrieves the user by phone hash, verifies the Argon2id hash, and issues a new token pair. `AuthService.RefreshToken()` performs one-time-use token rotation: the incoming raw token is verified against the stored Argon2id hash, the old token is marked as used, and a new token pair is issued. Refresh token reuse detection — a second use of an already-used token — triggers revocation of the entire token family for that user.

### Profile Module
`ProfileService.UpsertProfile()` validates display name length (max 60 chars), bio length (max 500 chars), and parses the birth date to compute the user's age using the exact formula: `age = now.Year() - birth.Year(); if now < birth.AddDate(age, 0, 0) { age-- }`. `ProfileService.UploadPhoto()` calls `mediaStore.UploadPhoto()` to write the image to MinIO under the object key `users/<userID>/photos/<uuid>.jpg`, records the media row in `social.media`, and updates the user's avatar URL. The `ProfileRepository.FindCandidates()` query uses `ST_DWithin` for geolocation filtering: `WHERE ST_DWithin(p.location, ST_SetSRID(ST_MakePoint($lon, $lat), 4326)::geography, $maxDistanceMetres)`.

### Matching Module
`MatchingService.GetCandidates()` executes a multi-step pipeline: (1) fetch requester settings from `settingsRepo`; (2) fetch seen IDs from `matchingCache.GetSeenIDs()`; (3) fetch rejected IDs from `matchRepo.GetRejectedIDs()`; (4) fetch blocked IDs from `matchRepo.GetBlockedIDs()`; (5) build `FindCandidatesOpts` with `ExcludeIDs` (union of seen, rejected, blocked sets) and `AllowedNiyyahs` (computed from requester's niyyah: nikah_year allows nikah_year + serious_marriage, serious_marriage allows both, friendship allows only friendship); (6) call `profileRepo.FindCandidates()` for the SQL query; (7) apply madhab +10 boost and no-photo blur in a loop; (8) add the returned IDs to the seen-set cache with 24-hour TTL. `MatchingService.Like()` calls `matchRepo.RecordLike()` — which performs a transactional UPSERT on `social.likes` and checks for mutual match — and on match creation sends an in-app notification and a non-blocking FCM push via the `pushCh chan<- domain.PushEvent` channel.

### Chat Module
The `Hub` struct is the central WebSocket connection registry. `Hub.ServeWS()` upgrades the HTTP connection to WebSocket, starts a 10-second authentication timeout goroutine, and waits for the first message. If the first message type is `"auth"`, it verifies the JWT token using `jwtManager.Verify()`, registers the connection in `connections[userID]`, cancels the timeout, and begins the `readPump` loop. The `readPump` goroutine receives messages, dispatches by type: `"chat_msg"` → decrypt-free storage via `chatSvc.SendMessage()` + Redis PUBLISH to `chat:<matchID>`; `"mahram_chat_msg"` → route via `mahramChatSvc`; `"typing"` → broadcast typing event to other user. A background Redis SUBSCRIBE goroutine receives messages from other Hub instances and delivers them to local connections. Panic recovery at the top of each read loop prevents Hub crash from individual connection failures.

### Feed Module
`PostService.CreatePost()` enforces the reputation gate: it calls `reputationSvc.GetScore()` and if the score is below 30 (or if `GetScore()` returns an error, which is treated as score=0 in the fail-closed implementation), the post creation is rejected with 403 Forbidden. Feed listing uses cursor-based pagination: `nextCursor` is the `created_at` timestamp of the last returned post, encoded as a base64 string.

### Reputation Module
The `TrustEngine` worker drains a `chan uuid.UUID` of user IDs requiring score recalculation. For each ID, it calls `trustGraphRepo.ComputeScore()`, which executes a Cypher query that retrieves all incoming `[:RATED]` edges, computes the Bayesian-smoothed weighted average, multiplies by 20, and returns the integer score. The score is written to `social.users.trust_score` (PostgreSQL) and `matchingCache.CacheTrustScore()` (Redis, 1-hour TTL).

---

## 5.4 Database Architecture

TrueConnect's database design reflects three guiding principles: relational integrity for matchmaking state, spatial support for geolocation discovery, and strict encryption for identity data.

**UUID Primary Keys** are used for all entities. This choice provides globally unique identifiers that can be generated at the application layer without a database round-trip, prevents enumeration attacks (sequential integer IDs expose the total number of users), and facilitates future data federation across multiple database instances.

**Like/Match Relationship Model.** The `social.likes` table has a `UNIQUE(liker_id, liked_id)` constraint that prevents duplicate like records. Mutual match detection is performed atomically within a database transaction in `match_repo.go`: when `RecordLike(userID, targetID)` is called, the transaction checks for the existence of the reverse like (`SELECT 1 FROM social.likes WHERE liker_id=$targetID AND liked_id=$userID`). If found, a match row is inserted into `social.matches` with `UNIQUE(LEAST(user_a_id, user_b_id), GREATEST(user_a_id, user_b_id))` (effectively normalising the pair order to prevent duplicate matches). If not found, only the like row is inserted.

**Message Storage.** Messages are stored as `bytea` ciphertext in `social.messages.content`. The AES-256-GCM nonce (12 bytes) is prepended to the ciphertext, allowing the decryption key and nonce to be reconstructed from the stored bytes. The `is_toxic` boolean flag supports content moderation without requiring message decryption at the query layer. Performance indexes: `idx_messages_match (match_id, created_at DESC)` for conversation retrieval; `idx_messages_sender (sender_id, created_at DESC)` for user-centric queries; a partial index `idx_messages_toxic (is_toxic, created_at DESC) WHERE is_toxic = true` for admin moderation queries.

**Geolocation.** Profile locations are stored as `geometry(Point, 4326)` (WGS84 coordinate system) in `social.profiles.location`. Discovery queries use `ST_DWithin(p.location, ST_SetSRID(ST_MakePoint($lon, $lat), 4326)::geography, $radiusMetres)` which leverages the PostGIS geography type for accurate great-circle distance calculations. A spatial index (`CREATE INDEX idx_profiles_location ON social.profiles USING GIST(location)`) allows the query planner to use a bounding-box filter before the precise distance calculation.

**Identity Vault.** The `identity_vault` schema is accessed only by a restricted PostgreSQL role (`vault_writer`), separate from the application role (`tc_app`) that owns `social.*`. This role separation means that even if the application-layer database credentials are compromised, the identity vault data (IIN, full name) cannot be accessed. All vault fields are AES-256-GCM encrypted; only the `iin_hash` (Argon2id) is stored unencrypted, enabling uniqueness checks without decryption.

**Key Indexes for Performance** (migration 000018): `idx_interactions_cooldown (rater_id, rated_id, created_at DESC)` for the 24-hour interaction rate limiting check; `idx_users_trust_score (trust_score DESC)` for leaderboard queries; `idx_notifications_user (user_id, is_read, created_at DESC)` for inbox loading; `idx_likes_liker (liker_id, liked_id)` for mutual match detection.

---

## 5.5 WebSocket Architecture

The WebSocket subsystem is implemented around the `Hub` struct in `internal/handler/chat_handler.go`. The Hub serves as both the connection registry and the message router.

**Connection lifecycle:**
1. Client sends `GET /v1/ws` with standard WebSocket upgrade headers
2. nginx forwards with `Upgrade: $http_upgrade` and `Connection: upgrade` headers; `proxy_read_timeout 86400s` keeps the connection alive
3. Hub calls `websocket.Accept()` with origin validation (`allowedOrigins` list, relaxed in dev mode when `isDev = true`)
4. A 10-second timeout goroutine is started. If no auth message arrives within 10 seconds, the connection is closed
5. Client sends `{"type":"auth","token":"<JWT access token>"}`
6. Hub calls `jwtManager.Verify(token)` to extract `userID`. On success, `connections[userID] = conn` (under `sync.RWMutex` write lock), timeout is cancelled
7. Hub sends `{"type":"auth_ok"}` to confirm
8. `readPump` goroutine begins: loop on `wsjson.Read()`, dispatch by `wsIncoming.Type`
9. On disconnect (any read error), connection is removed from the map and resources are released

**Message routing flow:**
- Sender → `readPump` goroutine → service call (encrypt + persist) → direct write to recipient connection (if present in `connections` map, under read lock) → Redis PUBLISH `chat:<matchID>` (always, for cross-instance delivery)
- Redis subscriber goroutine → receive from SUBSCRIBE → look up `connections[recipientID]` → write to local connection

**Panic recovery:** Each `readPump` goroutine has `defer func() { if r := recover(); r != nil { h.log.Error("websocket readPump panic", "error", r) } }()` at the top, preventing a panic in any single connection's message handler from crashing the Hub goroutine and disconnecting all other users.

**Flutter client reconnection:** `ChatNotifier` in `lib/providers/chat_provider.dart` implements exponential backoff: on connection error, it waits 1s, 2s, 4s, 8s, 16s before each retry attempt, up to a maximum of 5 attempts, after which the `isConnected` state is set to false and the user is shown a connection-lost indicator.

---

## 5.6 UML Diagram Specifications

The following sections provide complete textual specifications for all eight UML diagrams. Each specification is precise enough to draw the diagram without requiring additional information.

---

### Diagram A: Use Case Diagram

**Actors:**
1. **Muslim User** (primary) — seeks a spouse; can be male or female
2. **Mahram Guardian** (secondary) — supervises communication; is an authenticated user in the system
3. **System Administrator** (secondary) — manages platform health and content moderation

**Use Cases and relationships:**

Muslim User use cases:
- Register Account
- Login
- Declare Niyyah (`<<include>>` Login)
- Edit Profile
- Upload Profile Photo
- View Discovery Feed (`<<include>>` Authenticate)
- Like Candidate (`<<include>>` View Discovery Feed)
- Pass Candidate (`<<include>>` View Discovery Feed)
- View Matches
- Send Chat Message (`<<include>>` Authenticate, `<<include>>` Verify Match Exists)
- Send Mahram Chat Message (`<<extend>>` Send Chat Message — only when mahram room exists)
- Invite Mahram Guardian
- View Community Feed (`<<include>>` Authenticate)
- Create Post (`<<include>>` Trust Score ≥ 30)
- Like Post
- Add Comment
- Update Discovery Settings
- Submit KYC Documents
- Rate Interaction (`<<include>>` Verify Match Exists)
- View Leaderboard
- Report Content (Whisper)
- Block User
- View Imam Directory
- Confirm Nikah (`<<include>>` Verify Imam Selected)
- Mark Family Introduction Done

Mahram Guardian use cases (subset of Muslim User use cases plus):
- Join Mahram Chat Room
- Send Mahram Chat Message

System Administrator use cases:
- Review KYC Submissions
- Issue KYC Verdict (approve/reject)
- Review Sybil Clusters
- Review Whisper Flags
- List All Users

---

### Diagram B: System Architecture Diagram

**Layout:** Three vertical columns — Flutter Client (left), Go Backend + nginx (centre), Data Stores (right)

**Flutter Client column (top to bottom):**
- Presentation Layer: 21 screens (SplashScreen, OnboardingScreen, DiscoveryScreen, MatchesScreen, ChatScreen, MahramChatScreen, FeedScreen, ProfileScreen, etc.)
- State Layer: Riverpod providers (authStateProvider, matchingNotifierProvider, chatNotifierProvider, feedProvider, settingsNotifierProvider, etc.)
- Network Layer: DioClient (with _AuthInterceptor, _RetryInterceptor, _ErrorInterceptor) + WebSocketChannel
- Model Layer: User, Profile, Match, Post, UserSettings, AppNotification, Interaction, Venue

**Go Backend column (top to bottom):**
- nginx (port 80): rate limiting (api_limit 30/s, auth_limit 5/s), WebSocket upgrade headers, SSL/TLS termination
- Go API binary (port 8080): Gin router + middleware (auth, rate-limit, logging)
  - Handler layer: auth_handler, profile_handler, matching_handler, chat_handler (Hub), post_handler, settings_handler, mahram_handler, interaction_handler, kyc_handler, notification_handler, whisper_handler, imam_handler, admin_handler, user_handler
  - Service layer: AuthService, ProfileService, MatchingService, ChatService, PostService, SettingsService, MahramService, MahramChatService, InteractionService, ReputationService, KYCService, NotificationService, WhisperService, ImamService, AdminService
  - Repository interfaces (ports)
  - Adapter layer: postgres/, neo4j/, redis/, minio/
- Go Worker binary: TrustEngine goroutine, PushWorker goroutine, NiyyahTimerWorker goroutine, RefreshTokenCleanupTicker

**Data Stores column (top to bottom):**
- PostgreSQL 16 + PostGIS: social.* schema + identity_vault schema
- Neo4j 5 Community + GDS: (:User)-[:RATED]->(:User), (:User)-[:INTERACTED_WITH]->(:User)
- Redis 7: pub/sub channels (chat:<matchID>), seen-set keys, rate-limit counters, trust score cache
- MinIO: trueconnect bucket (users/*/photos/*.jpg, kyc/*/)

**Connections with labels:**
- Flutter ↔ nginx: HTTPS REST (port 443/80) + WSS upgrade (port 443/80)
- nginx → Go API port 8080: HTTP proxy (all /v1/* routes)
- nginx → Go API port 8080: WebSocket proxy (/v1/ws with Upgrade headers)
- Go API → PostgreSQL: pgx/v5 connection pool (port 5432)
- Go API → Neo4j: neo4j-go-driver v5 Bolt protocol (port 7687)
- Go API → Redis: go-redis/v9 (port 6379) — pub/sub + cache
- Go API → MinIO: minio-go v7 S3 API (port 9000)
- Go Worker → PostgreSQL, Neo4j, Redis: same drivers as API

---

### Diagram C: Sequence Diagram — User Registration and Onboarding

**Participants (left to right):** User, Flutter App, Dio HTTP Client, Go Auth Handler, AuthService, UserRepository (Postgres), JWT Manager, Secure Storage, Go Profile Handler, ProfileService, ProfileRepository

**Steps:**
1. User enters phone number and password on `RegisterScreen`
2. Flutter calls `AuthNotifier.register(phone, password)`
3. Dart sends `POST /v1/auth/register` via Dio with JSON body `{phone, password}`
4. Dio `_AuthInterceptor` passes through (no token for registration)
5. nginx routes to Auth Handler
6. `AuthHandler.Register()` binds and validates request body
7. `AuthHandler` calls `authSvc.Register(ctx, phone, password, email)`
8. `AuthService.Register()` calls `argon2id.GenerateFromPassword(password)` → `passwordHash`
9. `AuthService.Register()` calls `crypto.EncryptAES(phone)` → `phoneEncrypted`, `crypto.HashPhone(phone)` → `phoneHash`
10. `AuthService.Register()` calls `userRepo.Create(ctx, user)` → INSERT into `social.users`
11. `AuthService.Register()` calls `jwtManager.Sign(userID, isAdmin=false)` → `accessToken`
12. `AuthService.Register()` generates `refreshToken` (256-bit random), hashes it with Argon2id, calls `tokenRepo.Create(ctx, tokenHash, userID, expiresAt)` → INSERT into `social.refresh_tokens`
13. Response: `{access_token, refresh_token, user{id, verification_level}}`
14. Flutter stores `access_token` in FlutterSecureStorage key `"access_token"`
15. Flutter stores `refresh_token` in FlutterSecureStorage key `"refresh_token"`
16. Flutter caches user JSON in FlutterSecureStorage key `"cached_user"`
17. GoRouter `redirect()` detects `authState = User` → navigates to `/niyyah`
18. User selects niyyah on `NiyyahSelectionScreen` (e.g., `nikah_year`)
19. Flutter sends `PATCH /v1/profiles/me` with `{niyyah: "nikah_year"}`
20. `ProfileHandler.UpdateProfile()` calls `profileSvc.UpsertProfile(ctx, profile)`
21. `ProfileService.UpsertProfile()` calls `profileRepo.Upsert(ctx, profile)` → UPSERT into `social.profiles`
22. Response: `{status: "success", data: {profile}}`
23. Flutter navigates to `/home` (DiscoveryScreen)

---

### Diagram D: Sequence Diagram — Swipe Right and Match Creation

**Participants:** User A (Flutter), MatchingNotifier, Dio Client, nginx, Go Matching Handler, MatchingService, MatchRepository (Postgres), MatchingCache (Redis), NotificationService, PushChannel, User B (Flutter via WebSocket Hub)

**Steps:**
1. User A swipes right on a candidate card in `DiscoveryScreen`
2. Flutter calls `matchingNotifier.like(candidateUserID)`
3. `MatchingNotifier.like()` calls `POST /v1/matching/like` with body `{target_id: candidateUserID}`
4. Matching Handler validates JWT, extracts `userAID` from context
5. `MatchingHandler.Like()` calls `matchingSvc.Like(ctx, userAID, targetID)`
6. `MatchingService.Like()` validates `userAID != targetID` (returns `domain.ErrInvalidInput` if equal)
7. `MatchingService.Like()` calls `matchRepo.RecordLike(ctx, userAID, targetID)`
8. `RecordLike()` opens a PostgreSQL transaction
9. `RecordLike()` executes `INSERT INTO social.likes (liker_id, liked_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
10. `RecordLike()` checks `SELECT 1 FROM social.likes WHERE liker_id=$targetID AND liked_id=$userAID`
11a. **No reverse like found**: commits transaction, returns `(matched=false, matchID=uuid.Nil, nil)`
    → Response: `{matched: false}` → Flutter shows no banner
11b. **Reverse like found**: `RecordLike()` executes `INSERT INTO social.matches (user_a_id, user_b_id, matched_at, niyyah_timer_ends_at)`, commits transaction, returns `(matched=true, matchID, nil)`
12. (match case) `MatchingService.Like()` calls `notifSvc.Create(ctx, targetID, "match", ...)` → INSERT into `social.notifications`
13. (match case) `MatchingService.Like()` sends non-blocking to `pushCh <- domain.PushEvent{UserID: targetID, Title: "Жаңа мэтч!", ...}`
14. PushWorker receives from `pushCh`, fetches `user.FCMToken`, calls Firebase FCM API
15. (match case) `matchingCache.AddSeen(ctx, userAID, [targetID], 24h)` — adds target to seen set
16. Response: `{matched: true, match_id: "<uuid>"}` returned to Flutter
17. Flutter `MatchingNotifier` sets `_lastMatch = matchData`, triggers match banner display
18. User B receives FCM push notification (if app in background) or WebSocket `match_notification` event (if online)

---

### Diagram E: Sequence Diagram — Real-time Chat Message

**Participants:** User A (Flutter ChatScreen), ChatNotifier A, WebSocket Conn A, Hub (Go), ChatService, MessageRepository (Postgres), Redis pub/sub, Hub Instance B (optional), WebSocket Conn B, ChatNotifier B (Flutter)

**Steps:**
1. User A types a message and taps Send in `ChatScreen`
2. `ChatNotifier.send(content)` creates optimistic message `{id: "temp_<uuid>", content, sender_id: currentUserId, created_at: now}`, inserts into local `messages` list
3. `ChatNotifier` calls `channel.sink.add(jsonEncode({type: "chat_msg", payload: {match_id, content}}))`
4. WebSocket frame sent to Hub over TLS
5. Hub `readPump` goroutine receives `wsIncoming{type: "chat_msg", payload: wsChatPayload{match_id, content}}`
6. Hub calls `chatSvc.SendMessage(ctx, matchID, senderID, content)`
7. `ChatService.SendMessage()` verifies sender is a participant in this match: `matchRepo.IsMatched(ctx, senderID, peerID)`
8. `ChatService.SendMessage()` encrypts `content` with AES-256-GCM using `crypto.Encrypt([]byte(content))` → `ciphertext`
9. `ChatService.SendMessage()` calls `messageRepo.Create(ctx, message{match_id, sender_id, content: ciphertext})` → INSERT into `social.messages`
10. Hub looks up `connections[recipientID]` under read lock
11a. **Recipient connected locally**: Hub writes `wsOutgoing{type: "chat_msg", payload: {id, match_id, sender_id, content: plaintext, created_at}}` directly to recipient WebSocket connection
11b. **Recipient not connected locally** (or always, for cross-instance): Hub publishes to Redis: `PUBLISH chat:<matchID> <json_message>`
12. Redis delivers the PUBLISH to all subscribing Hub instances
13. Hub instance (may be same or different server) receives from SUBSCRIBE, looks up `connections[recipientID]`, writes to local connection
14. ChatNotifier B receives WebSocket message, checks `id` against existing messages to avoid duplicates
15. ChatNotifier B appends message to `state.messages` list; Flutter widget tree re-renders
16. Hub sends delivery confirmation back to Sender A with `{type: "chat_msg", id: <db_uuid>, ...}` replacing the optimistic `temp_*` entry: ChatNotifier A calls `indexWhere((m) => m['id'].startsWith('temp_') && m['content'] == content)` and replaces with confirmed message

---

### Diagram F: Sequence Diagram — JWT Token Refresh

**Participants:** Flutter ChatScreen (or any screen), Dio Client, _AuthInterceptor, Secure Storage, Go Auth Handler, AuthService, RefreshTokenRepository

**Steps:**
1. Dio sends any API request with `Authorization: Bearer <expired_access_token>`
2. Go middleware attempts to verify token → JWT expiry error → returns HTTP 401
3. `_AuthInterceptor.onError()` intercepts the 401 response
4. Interceptor checks `response.requestOptions.extra['retried'] != true` (prevents infinite retry loop)
5. Interceptor reads `refresh_token` from FlutterSecureStorage
6. Interceptor sends `POST /v1/auth/refresh` with body `{refresh_token: <raw_token>}` using a plain Dio instance (bypasses the interceptor to prevent recursion)
7. `AuthHandler.Refresh()` calls `authSvc.RefreshToken(ctx, rawToken)`
8. `AuthService.RefreshToken()` hashes the incoming token with Argon2id
9. `AuthService.RefreshToken()` calls `tokenRepo.GetByHash(ctx, hash)` → SELECT from `social.refresh_tokens`
10a. **Token not found or already used**: returns `domain.ErrNotFound` → 401 → interceptor clears all secure storage → `AuthNotifier` sets state to null → GoRouter redirects to `/auth/login`
10b. **Token valid and unused**: `AuthService.RefreshToken()` calls `tokenRepo.MarkUsed(ctx, tokenID)` → UPDATE `used_at = NOW()`
11. New access token generated via `jwtManager.Sign(userID, isAdmin)`
12. New refresh token generated (256-bit random), hashed and stored in `social.refresh_tokens`
13. Response: `{access_token: "<new_jwt>", refresh_token: "<new_raw_token>", user: {...}}`
14. Interceptor stores new `access_token` and `refresh_token` in FlutterSecureStorage
15. Interceptor retries the original failed request with `options.extra['retried'] = true` and the new Bearer token
16. Original request succeeds with HTTP 200; response forwarded to calling provider

---

### Diagram G: Entity-Relationship Diagram (ERD)

**Entities and relationships:**

**social.users** (1) ──── (1) **social.profiles** [FK: profiles.user_id → users.id CASCADE]

**social.users** (1) ──── (0..1) **social.user_settings** [FK: user_settings.user_id → users.id CASCADE]

**social.users** (1) ──── (∞) **social.likes** [FK: likes.liker_id → users.id CASCADE; likes.liked_id → users.id CASCADE]
Unique constraint: (liker_id, liked_id)

**social.users** × **social.users** (∞) ──── (∞) via **social.matches**
[FK: matches.user_a_id → users.id CASCADE; matches.user_b_id → users.id CASCADE]
Unique constraint: (LEAST(user_a_id, user_b_id), GREATEST(user_a_id, user_b_id))

**social.matches** (1) ──── (∞) **social.messages** [FK: messages.match_id → matches.id CASCADE]
**social.users** (1) ──── (∞) **social.messages** [FK: messages.sender_id → users.id CASCADE]

**social.users** (1) ──── (∞) **social.posts** [FK: posts.author_id → users.id CASCADE]
**social.posts** (1) ──── (∞) **social.comments** [FK: comments.post_id → posts.id CASCADE]
**social.users** (1) ──── (∞) **social.comments** [FK: comments.author_id → users.id CASCADE]
**social.posts** × **social.users** (∞) ──── (∞) via **social.post_likes** [Composite PK: (post_id, user_id)]

**social.users** (1) ──── (∞) **social.notifications** [FK: notifications.user_id → users.id CASCADE]

**social.users** (1) ──── (∞) **social.interactions** [FK: interactions.rater_id → users.id CASCADE; interactions.rated_id → users.id CASCADE]

**social.users** (1) ──── (∞) **social.mahrams** [FK: mahrams.woman_user_id → users.id CASCADE]

**social.matches** (1) ──── (0..1) **social.mahram_chat_rooms** [FK: mahram_chat_rooms.match_id → matches.id CASCADE]
**social.users** (1) ──── (∞) **social.mahram_chat_rooms** [FK: mahram_chat_rooms.mahram_user_id → users.id CASCADE]
**social.mahram_chat_rooms** (1) ──── (∞) **social.mahram_messages** [FK: mahram_messages.room_id → mahram_chat_rooms.id CASCADE]
**social.users** (1) ──── (∞) **social.mahram_messages** [FK: mahram_messages.sender_id → users.id CASCADE]

**social.users** (1) ──── (∞) **social.refresh_tokens** [FK: refresh_tokens.user_id → users.id CASCADE]
**social.users** (1) ──── (∞) **social.swipe_rejections** [FK → users.id CASCADE; Composite PK: (user_id, rejected_id)]
**social.users** (1) ──── (∞) **social.blocks** [FK → users.id CASCADE; Composite PK: (blocker_id, blocked_id)]
**social.users** (1) ──── (∞) **social.kyc_submissions** [FK: kyc_submissions.user_id → users.id CASCADE]
**social.users** (1) ──── (∞) **social.whisper_reports** [FK: reporter_id, reported_id → users.id CASCADE]

**social.users** (1) ──── (0..1) **identity_vault.iin_vault** [FK: iin_vault.user_id → social.users.id]

**Neo4j nodes (separate graph store):**
**(:User {uid})** ──[:RATED {rating, weight, created_at}]──> **(:User {uid})**
**(:User {uid})** ──[:INTERACTED_WITH {context, created_at}]──> **(:User {uid})**

---

### Diagram H: Class Diagram — Backend Service Interfaces

**Interface: ProfileRepository** (`internal/repository/profile_repo.go`)
```
+ Upsert(ctx context.Context, p *domain.Profile) error
+ GetByUserID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
+ FindCandidates(ctx context.Context, opts FindCandidatesOpts) ([]*CandidateRow, error)
+ GetLeaderboard(ctx context.Context, limit int) ([]domain.LeaderboardEntry, error)
+ SetMarriedViaApp(ctx context.Context, userID uuid.UUID) error
```

**Interface: MatchRepository** (`internal/repository/match_repo.go`)
```
+ RecordLike(ctx, userID, targetID uuid.UUID) (matched bool, matchID uuid.UUID, err error)
+ RecordPass(ctx, userID, targetID uuid.UUID) error
+ ListMatches(ctx, userID uuid.UUID, cursor string, limit int) ([]*domain.Match, nextCursor string, err error)
+ GetMatch(ctx, matchID, userID uuid.UUID) (*domain.Match, error)
+ IsMatched(ctx, userA, userB uuid.UUID) (bool, error)
+ Unmatch(ctx, matchID, userID uuid.UUID) error
+ UnmatchByUsers(ctx, userA, userB uuid.UUID) error
+ BlockUser(ctx, blockerID, blockedID uuid.UUID) error
+ GetBlockedIDs(ctx, userID uuid.UUID) ([]uuid.UUID, error)
+ GetRejectedIDs(ctx, userID uuid.UUID) ([]uuid.UUID, error)
+ MarkFamilyIntroDone(ctx, matchID uuid.UUID) error
+ MarkImamConfirmed(ctx, matchID uuid.UUID) error
+ FindExpiredNiyyahMatches(ctx context.Context) ([]*domain.Match, error)
+ ListMatchViews(ctx, userID uuid.UUID, cursor string, limit int) ([]*MatchViewRow, string, error)
+ GetPendingLikes(ctx, userID uuid.UUID) ([]uuid.UUID, error)
```

**Interface: TrustGraphRepository** (`internal/repository/trust_graph_repo.go`)
```
+ UpsertUser(ctx context.Context, userID uuid.UUID) error
+ RecordRating(ctx, raterID, ratedID uuid.UUID, rating int, weight float64) error
+ ComputeScore(ctx, userID uuid.UUID) (int, error)
+ DeleteUserNode(ctx, userID uuid.UUID) error
+ RunSybilDetection(ctx context.Context) ([]*SybilCluster, error)
```

**Struct: Hub** (`internal/handler/chat_handler.go`)
```
- mu sync.RWMutex
- connections map[uuid.UUID]*websocket.Conn
- chatSvc *service.ChatService
- matchSvc *service.MatchingService
- reputationSvc *service.ReputationService
- mahramChatSvc *service.MahramChatService
- rdb *redis.Client
- jwtManager *tcjwt.Manager
- log *slog.Logger
- allowedOrigins []string
- isDev bool
+ NewHub(...) *Hub
+ ServeWS(c *gin.Context)
- readPump(ctx, conn, userID)
- sendToUser(userID, msg)
- subscribeRedis(ctx)
```

**Struct: MatchingService** (`internal/service/matching_service.go`)
```
- profileRepo repository.ProfileRepository
- userRepo repository.UserRepository
- matchRepo repository.MatchRepository
- settingsRepo repository.SettingsRepository
- matchingCache repository.MatchingCache
- graphRepo repository.TrustGraphRepository
- notifSvc *NotificationService
- pushCh chan<- domain.PushEvent
+ NewMatchingService(...) *MatchingService
+ GetCandidates(ctx, requesterID uuid.UUID) ([]*CandidateView, error)
+ Like(ctx, userID, targetID uuid.UUID) (*LikeResult, error)
+ Pass(ctx, userID, targetID uuid.UUID) error
+ ListMatches(ctx, userID uuid.UUID, cursor string, limit int) ([]*MatchView, string, error)
+ GetGraphCandidates(ctx, userID uuid.UUID) ([]*CandidateView, error)
```

**Dependencies (arrows):**
- MatchingService ──uses──> ProfileRepository (interface)
- MatchingService ──uses──> MatchRepository (interface)
- MatchingService ──uses──> SettingsRepository (interface)
- MatchingService ──uses──> TrustGraphRepository (interface)
- Hub ──uses──> ChatService
- Hub ──uses──> MatchingService
- Hub ──uses──> MahramChatService
- postgres.ProfileRepo ──implements──> ProfileRepository
- postgres.MatchRepo ──implements──> MatchRepository
- neo4j.TrustGraphRepo ──implements──> TrustGraphRepository
