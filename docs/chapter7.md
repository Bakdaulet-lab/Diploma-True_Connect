# Chapter 7: Implementation and Deployment

## 7.1 Overview

This chapter documents the concrete implementation of TrueConnect from the first migration file to the production-ready Docker Compose stack. The project was built iteratively across fourteen sprints (Sprint 0 through Sprint 13), each producing a verifiable, tested increment. The backend reached feature completion at Sprint 11 with 102 passing service-layer tests; the Flutter mobile client reached feature completion at Sprint 13 with all 21 screens implemented. This chapter traces the implementation in four major areas: backend API and security infrastructure, real-time communication, frontend architecture and screens, and deployment pipeline.

---

## 7.2 Backend Implementation

### 7.2.1 Project Bootstrapping and Database Migrations

The backend was bootstrapped as a Go module (`github.com/trueconnect/backend`) using a strict Clean Architecture directory layout. The `cmd/` directory contains three entry points — `cmd/api/main.go` (HTTP server), `cmd/worker/main.go` (background jobs), and `cmd/migrate/main.go` (schema migrations) — enabling independent deployment of each component.

Database schema evolution is managed by `golang-migrate` through 19 versioned SQL migration files in `migrations/`. Each migration consists of a paired `.up.sql` and `.down.sql` file, enabling deterministic rollback. The migration sequence covers the full schema lifecycle:

- **Migration 000001** — `social` schema, `users` table with UUID primary key, `phone_hash` (bytea, unique), `phone_encrypted`, `email_encrypted`, `password_hash`, `public_key`, `verification_level` enum, `trust_score` int
- **Migration 000002–000005** — `profiles`, `likes`, `matches`, `messages` tables
- **Migration 000006–000008** — `posts`, `post_likes`, `comments` (social feed)
- **Migration 000009** — `identity_vault` schema, `iin_vault` table with AES-256-encrypted IIN and full name; `identity_vault_role` PostgreSQL role with GRANT restricted to vault schema only
- **Migration 000010–000012** — `kyc_submissions`, `notifications`, `user_settings`
- **Migration 000013–000015** — `niyyah`/`madhab`/`languages`/`no_photo_mode` columns added to profiles, `mahrams`, `mahram_chat_rooms`, `mahram_messages`, `whisper_reports` tables
- **Migration 000016–000018** — `refresh_tokens` (one-time-use rotation), `interactions`, `sybil_clusters`
- **Migration 000019** — `ON DELETE CASCADE` added to `interactions`, `matches`, `messages`, `whisper_reports` foreign keys

The Docker Compose service `migrate` runs `cmd/migrate/main.go` as a one-shot container on every `make docker-up`, with a `depends_on: postgres: condition: service_healthy` guard that prevents the API from starting before migrations complete.

### 7.2.2 Authentication Implementation

The authentication subsystem (`internal/service/auth_service.go`, `internal/handler/auth_handler.go`, `internal/adapter/postgres/user_repo.go`) implements a layered security model:

**Registration flow:**
1. The handler receives `{phone, email, password, gender, madhab}` and calls `auth_service.Register()`.
2. The service normalizes the phone number, computes `phone_hash = Argon2id(phone, salt)` for lookup without storing plaintext, and encrypts `phone_encrypted = AES-256-GCM(phone, ENCRYPTION_KEY, random_nonce)` for recovery display.
3. `password_hash = Argon2id(password, salt)` with parameters 64 MB memory, 3 iterations, 4 parallelism threads — a profile chosen to require ~250 ms on a single CPU core, raising the cost of offline dictionary attacks to prohibitive levels.
4. The user row is inserted; a profile row with default values is created in the same transaction.
5. A 15-minute JWT access token and 7-day refresh token are returned; the refresh token's SHA-256 hash is stored in `social.refresh_tokens` alongside `expires_at`.

**Token refresh and family revocation:**
The refresh endpoint (`POST /v1/auth/refresh`) validates the incoming token against the stored hash, marks the token as `used_at = NOW()`, and issues a new pair. If a refresh token that has already been marked `used_at` is presented (indicating token theft and reuse), the entire token family for that user is immediately revoked — all rows in `social.refresh_tokens` for `user_id` are deleted. This is the OWASP-recommended refresh token rotation with reuse detection.

**PII field encryption:**
All personally identifiable fields — `phone_encrypted`, `email_encrypted`, and the `identity_vault` columns `iin_encrypted` and `full_name_encrypted` — use AES-256-GCM with a 12-byte random nonce prepended to the ciphertext. The nonce is generated via `crypto/rand.Read` on each encryption call, ensuring that two encryptions of the same plaintext produce different ciphertexts and that nonce reuse (which would break GCM's authentication guarantee) is computationally infeasible.

### 7.2.3 Profile and Media Implementation

The profile subsystem stores biographical, preference, and location data in `social.profiles` with two geographic columns: `city` (text) and `location` (PostGIS GEOGRAPHY type, SRID 4326). Avatar and photo uploads go through the MinIO adapter (`internal/adapter/minio/`): the handler accepts a `multipart/form-data` request, validates the MIME type against an allowlist (`image/jpeg`, `image/png`, `image/webp`), and uploads to the `avatars` or `media` MinIO bucket, returning a presigned URL stored in `profiles.avatar_url`.

The `no_photo_mode` boolean allows female users to hide their avatar from the discovery feed. When `no_photo_mode = true`, the profile adapter returns `AvatarURL = ""` and sets a flag `AvatarBlurred = true` in the `CandidateView` struct, which the Flutter client interprets to display a blurred silhouette placeholder rather than an empty image widget.

### 7.2.4 Matching Algorithm Implementation

The candidate discovery query (`GET /v1/matching/candidates`) executes five filtering stages in sequence:

1. **Seen-set exclusion**: A Redis set `seen:<userId>` is loaded; the query excludes any `user_id` in this set.
2. **Block exclusion**: `social.blocks` table is left-joined; profiles with an active block in either direction are excluded.
3. **Geospatial filter**: `ST_DWithin(p.location, ST_MakePoint($lon, $lat)::geography, $radius_m)` applied using the user's current latitude/longitude and `user_settings.max_distance_km`.
4. **Preference filters**: Age range (`min_age` / `max_age`), `niyyah_filter` (requires matching niyyah enum value when set), `madhab_filter` (requires matching madhab when set), and `show_me` (gender preference).
5. **Ordering**: Candidates are sorted by `trust_score DESC` then by same-madhab match (+10 virtual score boost applied in Go after the query rather than in SQL to avoid polluting the stored score).

The returned batch of 20 candidates is loaded into the Flutter `matchingNotifierProvider` which renders them as a swipeable card stack on `DiscoveryScreen`. Each swipe right calls `POST /v1/matching/like`; the service atomically checks for a reciprocal like and creates a match if found, triggering a `match_notification` WebSocket push to both users.

**Niyyah timer:** When a match is created, `social.matches.niyyah_timer_ends_at` is set to `NOW() + 90 days`. The `NiyyahTimerWorker` runs daily as a background job in `cmd/worker/main.go`, querying matches where `niyyah_timer_ends_at < NOW()` and neither `family_intro_done = true` nor `imam_confirmed = true`, and sending a push notification reminder to both users. The Flutter `Match` model exposes `daysLeftOnTimer = max(0, niyyahTimerEndsAt.difference(DateTime.now()).inDays)` as a computed property, displayed as a colour-coded chip on the match card (green > 30 days, orange 8–30 days, red ≤ 7 days).

### 7.2.5 Trust Engine Implementation

The trust engine is implemented across three files: `internal/worker/trust_engine.go` (orchestration), `internal/service/reputation_service.go` (business logic), and `internal/adapter/neo4j/trust_repo.go` (graph queries).

When a user submits a post-meeting interaction rating (`POST /v1/interactions`), the service:
1. Inserts the interaction into `social.interactions` with `confirmed_at = NULL`.
2. Enqueues a Neo4j write via a Go channel to the `TrustEngine` worker.
3. The worker's `recalculateTrustScore()` function calls the Neo4j GDS PageRank projection, projecting nodes `:User` and relationships `:RATED` with the interaction weight as the relationship property.
4. The resulting score for the rated user is read from the GDS projection, Bayesian-smoothed toward a 2.5/5 neutral baseline with `C = 5` virtual ratings, and scaled to 0–100.
5. The score is written atomically to three stores: Neo4j (canonical node property), `social.users.trust_score` (for SQL joins), and `Redis key trust:<userId>` (for sub-millisecond swiping feed reads).

**Sybil detection** runs every 6 hours via a `time.Ticker` in the worker. The Louvain community detection algorithm (`gds.louvain.stream`) is projected over the full `:User`–`:INTERACTED_WITH` graph. Clusters where more than 70% of edges are internal (low external connectivity) and the KYC verification rate is below 20% are written to `social.sybil_clusters` and flagged for admin review at `GET /v1/admin/sybil/pending`.

### 7.2.6 Real-Time Chat Implementation

The WebSocket hub is implemented in `internal/handler/chat_handler.go`. Key implementation details:

**Hub struct:**
```
Hub {
    connections  map[uuid.UUID]*websocket.Conn   // goroutine-safe via mu
    mu           sync.RWMutex
    chatService  ChatService
    matchService MatchService
    mahramService MahramService
    contentFilter ContentFilter
    pushService  PushService
    jwtManager   *jwt.Manager
    allowedOrigins []string
    isDev        bool
}
```

**Connection lifecycle:**
1. The HTTP handler performs an origin check against `allowedOrigins` (loaded from `CORS_ORIGINS` env var); in dev mode all origins are allowed.
2. `coder/websocket` upgrades the HTTP connection.
3. A goroutine is launched for the connection; the hub's `mu.Lock()` registers the `conn` in the `connections` map.
4. The first incoming message must be `{"type":"auth","token":"<JWT>"}` within 10 seconds; a `context.WithTimeout` enforces the deadline.
5. On successful auth, `{"type":"auth_ok"}` is sent and message handling begins.
6. On connection close (graceful or error), `mu.Lock()` removes the entry from the map.
7. A `defer func() { recover() }()` wraps each goroutine's main loop to prevent a panicking connection from crashing the entire Hub.

**Message routing:**
Incoming `chat_msg` payloads are: (1) sanitized (HTML-stripped via `internal/pkg/sanitize`), (2) checked by the content filter (toxic classifier), (3) AES-256-GCM encrypted and stored in `social.messages`, (4) published to Redis channel `chat:<matchId>`. All Hub instances subscribe to `chat:*` and deliver matching messages to their locally connected users. If the recipient is not connected to any instance, the push service sends an FCM notification via Firebase.

**Mahram supervision:** When a `mahram_chat_msg` is received, the Hub verifies that the sender is a participant in the mahram room (via `mahramService.IsParticipant()`), encrypts and stores in `social.mahram_messages`, and delivers to all three room participants (woman, man, guardian).

### 7.2.7 Mahram, Imam, and KYC Features

**Mahram registration** (`POST /v1/mahram`) accepts a guardian's phone number, stores `mahram_phone_hash = Argon2id(phone, salt)` in `social.mahrams.mahram_phone_hash`. When the guardian registers or logs in from that phone number, the system resolves the hash and grants them access to the associated mahram rooms. This approach avoids storing the guardian's plaintext phone number while still enabling lookup by phone at login time.

**Imam catalog** (`GET /v1/imams`) is served from an embedded JSON catalog compiled into the binary via Go's `//go:embed` directive in `internal/pkg/imam/`. The catalog contains 10 imams across 5 Kazakhstan cities (Almaty, Astana, Shymkent, Aktobe, Karaganda). Because this data is immutable reference data, embedding it eliminates a database round-trip for every discovery screen visit. The `GET /v1/venues` endpoint similarly serves a hardcoded list of halal-certified meeting venues.

**Nikah confirmation** (`POST /v1/matches/:id/nikah-confirm`) sets `social.matches.imam_confirmed = true` and `social.profiles.married_via_app = true` for both users, locking the match and triggering a congratulatory push notification with the Quranic barakah phrase "Baraka Allahu feekum."

**KYC flow:** The `POST /v1/kyc/submit` endpoint accepts a `multipart/form-data` upload. The handler validates the file MIME type (allowlist: `image/jpeg`, `image/png`, `application/pdf`), uploads the document to MinIO under a UUID-keyed path in the `kyc-documents` bucket (not publicly accessible), and inserts a row into `social.kyc_submissions` with status `pending`. The admin review endpoint (`POST /v1/admin/kyc/:userId/verdict`) accepts `approved` or `rejected`; on approval, `social.users.verification_level` is updated to `id_verified`, triggering the 1.5× trust weight on future interaction ratings.

---

## 7.3 Frontend Implementation

### 7.3.1 Architecture and State Management

The Flutter application (`frontend/`) follows a feature-first directory structure:

```
frontend/lib/
├── core/
│   ├── constants/    # api_constants.dart, app_colors.dart
│   ├── router/       # app_router.dart (GoRouter + auth guard)
│   ├── theme/        # app_theme.dart, app_spacing.dart, app_shadows.dart
│   └── widgets/      # trust_score_badge.dart, niyyah_badge.dart, madhab_badge.dart
├── features/
│   ├── auth/         # login_screen, register_screen, auth_provider
│   ├── discovery/    # discovery_screen, matching_provider
│   ├── chat/         # chat_screen, mahram_chat_screen, chat_provider
│   ├── feed/         # feed_screen, create_post_screen, feed_provider
│   ├── matches/      # matches_screen, matches_provider
│   ├── profile/      # profile_screen, edit_profile_screen, profile_detail_screen
│   ├── settings/     # settings_screen, settings_provider
│   ├── notifications/# notifications_screen, notifications_provider
│   ├── kyc/          # kyc_screen, kyc_provider
│   └── imam/         # imam_connect_screen
└── models/           # user.dart, profile.dart, match.dart, post.dart, ...
```

All state management uses **Riverpod 2.x**. Providers are defined at the feature level and consumed via `ref.watch()` / `ref.read()` in widgets. The `authStateProvider` (`AsyncNotifierProvider<AuthNotifier, User?>`) is the root dependency: GoRouter's `_AuthRouterNotifier` watches it and redirects unauthenticated users to `/auth/login` on every navigation event.

### 7.3.2 HTTP Client and Interceptors

`DioClient` (`core/network/dio_client.dart`) wraps the `Dio` HTTP client with three stacked interceptors:

1. **Auth interceptor** — reads the JWT access token from `flutter_secure_storage` and attaches it as `Authorization: Bearer <token>` on every outgoing request. On a 401 response, the interceptor calls `POST /v1/auth/refresh` with the refresh token (stored separately in secure storage), stores the new token pair, and retries the original request transparently.

2. **Retry interceptor** — retries failed requests (network errors, 5xx responses) up to three times with exponential backoff: 1 s, 2 s, 4 s. Requests are not retried on 4xx client errors.

3. **Error interceptor** — maps HTTP error codes to user-facing snackbar messages: 422 → validation error details extracted from the response body; 429 → "Too many requests, please slow down"; 5xx → "Server error, please try again."

The base URL and all endpoint paths are defined in `core/constants/api_constants.dart` as compile-time `static const String` values. The `fixImageUrl()` helper replaces the `localhost` host in MinIO presigned URLs with the configured server host, enabling images to load correctly on physical devices that cannot resolve `localhost`.

### 7.3.3 WebSocket Client

The `chatNotifierProvider` manages the WebSocket lifecycle:

1. `connect(matchId)` opens a `WebSocketChannel` to `ws://<host>:8080/v1/ws`.
2. Immediately sends `{"type":"auth","token":"<JWT>"}` as the authentication handshake.
3. Subscribes to `channel.stream` via `stream.listen(_handleMessage, onError: _handleError, onDone: _handleDone)`.
4. `_handleMessage()` decodes the JSON envelope and dispatches on `type`: `chat_msg` → appends to message list and triggers `setState`; `typing` → sets `isTyping = true` with a 3-second reset timer; `content_warning` → shows a yellow banner; `content_blocked` → removes the last optimistic message from the list; `match_notification` → shows a congratulation overlay.
5. On `onDone` (connection closed), the reconnect loop fires: exponential backoff delays of 1 s, 2 s, 4 s, 8 s, 16 s (maximum 5 attempts), then gives up and shows an offline banner with a manual "Reconnect" button.
6. `send(matchId, text)` serialises `{"type":"chat_msg","payload":{"match_id":"<id>","content":"<text>"}}` and writes to `channel.sink`. An optimistic message is appended to the state list immediately; a `content_blocked` response removes it.

Typing indicators are sent via `sendTyping(matchId)` which is debounced: repeated keystrokes within 300 ms collapse to a single typing message, preventing the server from receiving hundreds of events per second on fast typists.

### 7.3.4 Key Screen Implementations

**DiscoveryScreen (`/home`):** Renders the candidate stack using a gesture detector that tracks horizontal drag velocity. A rightward fling above the threshold triggers `matchingNotifier.like(candidateId)`; leftward triggers `pass(candidateId)`. The card displays: hero-animated profile photo (or blurred silhouette for `noPhotoMode`), display name + age, `TrustScoreBadge` (animated counter widget), `NiyyahBadge` and `MadhabBadge` chips, and the city name. A filter sheet (bottom sheet) exposes sliders for max distance, age range, and dropdowns for niyyah filter and madhab filter, wired to `settingsNotifierProvider.update()` with a 450 ms debounce.

**ChatScreen (`/chat/:matchId`):** Displays messages in a `ListView.builder` with `reverse: true` (newest at bottom). Each message bubble uses the sender's trust score colour: own messages are displayed in `AppColors.gold` bubbles on the right; partner's messages in `AppColors.surface` on the left. The text input row contains a send button and, for female users in a mahram-capable match, a guardian invite button that opens a bottom sheet for mahram room creation. A yellow `ContentWarningBanner` slides in when a `content_warning` WS message is received; a typing indicator ("Алихан пишет...") appears when `isTyping = true`.

**MahramChatScreen (`/mahram-chat/:roomId`):** Three-column colour coding: woman's messages in `AppColors.green`, man's messages in `AppColors.blue`, guardian's messages in `AppColors.gold`. A pinned banner at the top reads "Этот чат находится под надзором махрама" (This chat is supervised by your mahram) in Amiri Arabic script below the Kazakh text. The same `chatNotifierProvider` handles this screen using a different `roomId` parameter; the server routes `mahram_chat_msg` types to the mahram room rather than the match room.

**SettingsScreen (`/settings`):** Uses a `Column` of grouped settings tiles. The "Mahram" section shows the registered guardian's obfuscated phone (last 4 digits shown), with an "Add Mahram" text field and a delete button. The niyyah filter dropdown maps to the three `Niyyah` enum values using a localised label function. The madhab filter uses a `DropdownButtonFormField` with the five `Madhab` enum values. All changes call `settingsNotifierProvider.update()` with a 450 ms debounce followed by a PATCH to `GET /v1/settings`; on error the state is reverted to its previous value (optimistic update reversal).

**ImamConnectScreen (`/imams`):** A city `DropdownButton` filters the imam list. Each imam card shows name, specialisation, phone number, and a "Confirm Nikah" button. Tapping the button opens a confirmation dialog with the match partner's name; on confirmation, calls `POST /v1/matches/:id/nikah-confirm` and shows a full-screen congratulation overlay with the Quranic dua and an animated 8-pointed star.

**KycScreen (`/kyc`):** Uses `image_picker` to capture or select an identity document photo. The selected image is displayed in a preview card; a submit button calls `DioClient.post('/kyc/submit', formData: FormData.fromMap({'document': MultipartFile.fromFileSync(path)}))`. On success, an animated checkmark is displayed and the screen transitions to a "Verification Pending" state. `kycStatusProvider` polls `GET /v1/kyc/status` every 30 seconds while in `pending` status, automatically transitioning to `approved` or `rejected` without requiring a manual page refresh.

### 7.3.5 Islamic Design System (Дала Нұры)

The design system is implemented in `core/theme/` and `core/widgets/`:

- **AppColors**: `gold = #C9A84C`, `surface = #1A1A2E`, `background = #0F0F1E`, `green = #2ECC71`, `blue = #3498DB`, `whisper = #95A5A6`
- **AppTheme**: dark-first Material 3 theme with Nunito for Latin text and Amiri for Arabic/Quranic text, applied via `google_fonts` package
- **HalalPatternPainter**: a `CustomPainter` that tiles the Қошқар мүйіз (ram's horns) scrollwork pattern across the background canvas using `Path.addArc` and `Path.cubicTo` curves
- **IslamicStarWidget**: an 8-pointed star drawn via `CustomPainter` by rotating 8 isoceles triangles around a centre point, animated via `AnimationController` from `opacity 0.0` to `1.0` over 800 ms on the splash screen
- **KazakhDivider**: a `CustomPainter` that draws a repeating traditional Kazakh border motif using `Path.moveTo` and `Path.lineTo`, used as a section separator throughout the app
- **TrustScoreBadge**: displays the 0–100 trust score as an `AnimatedCounter` that counts up from 0 to the actual value over 600 ms on first render, with a ring colour that transitions from red (< 40) to orange (40–70) to green (> 70) via `ColorTween`
- **NiyyahBadge / MadhabBadge**: pill-shaped chips with localised Kazakh labels and a subtle gold border, displayed on both candidate cards and profile detail screens

---

## 7.4 Security Implementation

### 7.4.1 Transport Security

All API traffic flows through nginx, which enforces:
- `X-Content-Type-Options: nosniff` — prevents MIME-type sniffing attacks
- `X-Frame-Options: DENY` — prevents clickjacking via iframe embedding
- `X-XSS-Protection: 1; mode=block` — legacy XSS filter for older browsers
- `client_max_body_size 10m` — prevents memory exhaustion from oversized request bodies
- Rate limiting via `limit_req_zone $binary_remote_addr zone=api_limit:10m rate=30r/s` with `burst=20` for general API endpoints, and a separate `auth_limit` zone at `rate=5r/s burst=5` for authentication endpoints

### 7.4.2 Input Validation and Sanitisation

All incoming request bodies are decoded into typed Go structs with `go-playground/validator/v10` tags (e.g., `validate:"required,min=8,max=64"` for passwords, `validate:"e164"` for phone numbers). Validation errors are mapped to structured 422 responses listing each failing field. All user-generated text content (post content, chat messages, bio text) is passed through `internal/pkg/sanitize.StripHTML()` before storage, preventing stored XSS attacks from rendering malicious HTML in the Flutter `Text` widget via `flutter_html` or similar packages.

### 7.4.3 SQL Injection Prevention

Every database query in the adapter layer uses parameterised queries via `pgx/v5`'s `$1`, `$2`, ... placeholder syntax. No string concatenation is used for query construction at any point in the codebase. The `go vet` and `golangci-lint` steps in CI would flag `fmt.Sprintf`-constructed queries if accidentally introduced.

### 7.4.4 CORS and Origin Validation

CORS allowed origins are loaded from the `CORS_ORIGINS` environment variable as a comma-separated list. The Gin middleware (`internal/handler/middleware/cors.go`) validates the `Origin` request header against this list; unlisted origins receive a 403. The WebSocket hub performs the same origin check: if `Hub.isDev = false`, connections from origins not in `Hub.allowedOrigins` are rejected before the HTTP Upgrade. In the Docker Compose environment, `isDev = (APP_ENV == "development")`, allowing all origins during local development without modifying the allowlist.

### 7.4.5 Content Moderation

The content filter (`internal/pkg/contentfilter/`) implements a keyword-based toxic message classifier that checks both Arabic-script and Latin-script patterns. Messages classified as `is_toxic = true` are stored in `social.messages` with the `is_toxic` flag set and a `content_blocked` WebSocket event is sent to the sender; the recipient never receives the message. Borderline messages receive a `content_warning` event, informing the sender that the message will be reviewed. All decisions are logged with the message ID for admin audit.

---

## 7.5 Deployment Pipeline

### 7.5.1 Docker Compose Stack

The production stack is defined in `deployments/docker-compose.yml` and brings up seven services:

| Service | Image | Port | Notes |
|---------|-------|------|-------|
| postgres | `postgis/postgis:16-3.4-alpine` | 5433:5432 | PostGIS extension pre-installed |
| neo4j | `neo4j:5-community` | 7474, 7687 | GDS plugin mounted via volume |
| redis | `redis:7-alpine` | 6379 | `maxmemory 256mb`, `allkeys-lru` |
| minio | `minio/minio:latest` | 9000, 9001 | S3-compatible object storage |
| migrate | Custom (`cmd/migrate`) | — | One-shot; `depends_on postgres healthy` |
| api | Custom (`cmd/api`) | 8080 | `depends_on migrate` |
| nginx | `nginx:alpine` | 80:80 | Reverse proxy + rate limiting |

The API Dockerfile uses a two-stage multi-stage build: the builder stage (`golang:1.25-alpine`) compiles with `CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" ./cmd/api` to produce a stripped static binary; the runtime stage (`alpine:3.19`) copies only the binary, creating a ~12 MB final image. The container runs as non-root user `appuser` (UID 1000) created via `adduser -D -u 1000 appuser`.

### 7.5.2 Nginx Configuration

The nginx configuration (`deployments/nginx/nginx.conf`) serves two roles: rate limiting and WebSocket proxying.

**Rate limiting zones:**
```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m  rate=30r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=5r/s;
```

Authentication endpoints (`/v1/auth/`) use `auth_limit` (5 req/s, burst 5), tightly constraining brute-force login attempts. All other API endpoints use `api_limit` (30 req/s, burst 20).

**WebSocket proxy:**
```nginx
location /v1/ws {
    proxy_pass http://api_backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400s;
    proxy_send_timeout 86400s;
}
```

The 86400-second (24-hour) read and send timeouts prevent nginx from terminating idle WebSocket connections between chat sessions. Without these settings, nginx would close connections after the default 60-second timeout, forcing clients to reconnect during any pause in a conversation.

### 7.5.3 Environment Configuration

All secret and environment-specific values are loaded from environment variables at startup. The `cmd/api/main.go` `loadConfig()` function reads and validates:

- `JWT_SECRET` — must be ≥ 32 characters; application panics at startup if missing
- `ENCRYPTION_KEY` — must be exactly 32 bytes hex-encoded (64 hex chars); panics if wrong length
- `FIREBASE_CREDENTIALS` — path to Firebase service account JSON; falls back to a `MockPushProvider` (logs pushes to stdout) if not set, enabling local development without Firebase credentials
- All database connection strings (`PG_HOST`, `NEO4J_URI`, `REDIS_HOST`, etc.) with defaults suitable for the Docker Compose service names

The `.env.example` file documents every required variable with comments. `make docker-up` reads `.env` via `env_file:` in the Compose file.

### 7.5.4 Continuous Integration

The CI pipeline is defined in `.github/workflows/ci.yml` and runs on every push and pull request to `main`:

```yaml
jobs:
  build-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - name: Build API
        run: go build ./cmd/api/...
      - name: Build Worker
        run: go build ./cmd/worker/...
      - name: Test
        run: go test ./internal/... -race -count=1
      - name: Vet
        run: go vet ./...
```

The `-race` flag instruments the binary with Go's race detector, which dynamically checks for concurrent map reads/writes and unsynchronised variable accesses. The 102 service-layer tests pass cleanly under the race detector. The `-count=1` flag disables test result caching, ensuring every CI run executes all tests rather than replaying cached results.

### 7.5.5 Makefile Developer Experience

The `Makefile` at the repository root provides a unified interface for all common operations:

| Target | Command | Purpose |
|--------|---------|---------|
| `build` | `go build ./cmd/...` | Compile all three binaries |
| `run` | `go run ./cmd/api/` | Start API with hot reload (air if installed) |
| `run-worker` | `go run ./cmd/worker/` | Start background worker |
| `test` | `go test ./internal/... -race` | Run all tests with race detector |
| `lint` | `golangci-lint run` | Static analysis |
| `migrate-up` | `cmd/migrate up` | Apply all pending migrations |
| `migrate-down` | `cmd/migrate down 1` | Rollback last migration |
| `docker-up` | `docker-compose up --build -d` | Build and start full stack |
| `docker-down` | `docker-compose down` | Stop all services |
| `docker-reset` | `docker-compose down -v && docker-compose up --build -d` | Full reset including volumes |
| `docker-logs` | `docker-compose logs -f` | Tail all service logs |
| `seed-halal` | `psql < migrations/seed_halal_demo.sql` | Load demo users |
| `reset-demo` | `migrate-down + docker-reset + seed-halal` | Clean demo environment |

The `seed-halal` target loads a reproducible demo dataset (`migrations/seed_halal_demo.sql`) containing two Almaty users — Айгерим Бекова (female, niyyah: nikah_year, madhab: hanafi) and Алихан Сейткали (male, niyyah: serious_marriage, madhab: hanafi) — with a pre-created match, a mahram room, and sample messages, enabling instant end-to-end demonstration without manual registration.

---

## 7.6 Testing Summary

**Table 7.1 — Test Coverage by Sprint**

| Sprint | Feature Tested | New Tests | Cumulative |
|--------|---------------|-----------|------------|
| 1 | Auth (register, login, refresh, logout, brute-force) | 30 | 30 |
| 2 | Profiles CRUD, PostGIS candidates, like/pass, settings | 21 | 51 |
| 3 | Interactions, trust score, Sybil detection, worker | 19 | 70 |
| 4 | WebSocket chat, feed CRUD, KYC stub, message encryption | 25 | 95 |
| 5 | Token family revocation, rate limiting, reports | 4 | 99 |
| 6–8 | Mahram, whisper, settings adapters | 19 | 118 |
| 9 | Halal filters (niyyah, madhab, no-photo) | 0 (integration) | 118 |
| 10–11 | Mahram chat, family intro, imam catalog | 7 | 120 (service) + imam |

All 102 active service-layer tests pass with `-race -count=1`. The test suite uses mock repository implementations that satisfy the repository interfaces — for example, `MockMatchRepository` implements all eight methods of `repository.MatchRepository`, enabling `MatchingService` tests to run without a live PostgreSQL connection. This approach was validated against the real adapters during integration testing via `make docker-up`.

---

## 7.7 Demo Walkthrough

The complete end-to-end demo scenario, reproducible via `make docker-up && make seed-halal`, follows this path:

1. **Launch** — `flutter run -d android` opens the splash screen with the animated 8-pointed Islamic star and the Quranic verse Ar-Rum 30:21.
2. **Onboarding** — three slides with Қошқар мүйіз CustomPainter background; "Начать" navigates to NiyyahSelection.
3. **Login** — enter Айгерим's demo credentials; `authStateProvider` calls `POST /v1/auth/login`, stores tokens in `flutter_secure_storage`, routes to `/home`.
4. **Discovery** — Алихан's card appears at the top of the stack (PostGIS candidate query within 50 km, same hanafi madhab → +10 boost). Swipe right triggers `POST /v1/matching/like`; the server detects a mutual like (Алихан was pre-seeded to have liked Айгерим), creates a match, and sends a `match_notification` WS event.
5. **Match banner** — a gold congratulation overlay slides up with both users' avatars.
6. **Chat** — navigate to `/chat/<matchId>`; Дала Нұры bubble design; type a message → real-time delivery.
7. **Mahram invite** — tap the guardian invite button; enter the guardian's phone number → `POST /v1/mahram`; guardian joins and the chat transitions to 3-way MahramChatScreen.
8. **Settings** — navigate to `/settings`; adjust age range; observe the discovery feed update on return.
9. **KYC** — navigate to `/kyc`; upload a document photo; status shows "Pending" then transitions to "Approved" after admin verdict.
10. **Imam Connect** — navigate to `/imams?matchId=<id>`; select Almaty; tap an imam card → confirm nikah; "Baraka Allahu feekum" success overlay with animated star.

This scenario exercises all 15 backend modules, all 50 API endpoints, the WebSocket hub, Redis pub/sub, Neo4j trust score, PostGIS candidate query, MinIO upload, and all 21 Flutter screens.
