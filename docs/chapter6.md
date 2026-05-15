# Chapter 6: Technology Selection and Justification

## 6.1 Backend Language and Runtime

### 6.1.1 Candidates Evaluated

The selection of a backend programming language is among the most consequential architectural decisions for any long-lived distributed system. Three candidates were evaluated in depth for TrueConnect: **Go (Golang) 1.25**, **Node.js 20 LTS (TypeScript)**, and **Python 3.12 (FastAPI/asyncio)**. Each represents a credible contemporary choice; the evaluation criteria were concurrency model, memory footprint, type safety, deployment size, ecosystem maturity for the required libraries, and operational characteristics on a single Kazakhstan-resident VPS.

### 6.1.2 Comparative Analysis

**Table 6.1 — Backend Language Comparison**

| Criterion | Go 1.25 | Node.js 20 + TS | Python 3.12 + FastAPI |
|-----------|---------|-----------------|----------------------|
| Concurrency model | CSP goroutines (M:N scheduler, 2KB initial stack) | Event loop + Worker Threads | asyncio (single-threaded event loop) |
| Memory per idle connection | ~4 KB (goroutine stack) | ~20-40 KB (V8 heap overhead) | ~40-80 KB (Python object overhead) |
| Compiled binary size | ~8 MB static (CGO_ENABLED=0) | N/A (runtime required) | N/A (interpreter required) |
| Cold-start latency | <5 ms | ~80-120 ms (V8 JIT warm-up) | ~200-400 ms (import time) |
| Type safety | Static, checked at compile time | Optional (TypeScript transpiled) | Optional (type hints, runtime mypy) |
| Race condition detection | Built-in `-race` detector | Limited (single-threaded mitigates but Worker Threads expose races) | GIL prevents true parallelism for CPU tasks |
| PostgreSQL driver | `pgx/v5` (native wire protocol, prepared statements) | `pg` / `pg-promise` | `asyncpg` / `psycopg3` |
| Neo4j driver | `neo4j-go-driver/v5` (official) | `neo4j-driver` (official) | `neo4j` (official) |
| WebSocket | `coder/websocket` (idiomatic goroutine per connection) | `ws` / `socket.io` | `websockets` / Starlette |
| Docker image size | ~12 MB (alpine + static binary) | ~180 MB (node:20-alpine + node_modules) | ~200 MB (python:3.12-slim + venv) |
| Deployment complexity | Single binary + env vars | npm install + transpile step | pip install + uvicorn |
| CI build time (approx.) | ~25 s (`go build`) | ~60 s (tsc + esbuild) | ~30 s (no compile, but slow test imports) |

**Concurrency under load.** TrueConnect's WebSocket hub maintains one goroutine per active WebSocket connection. In Go, goroutines are multiplexed onto OS threads by the runtime scheduler; each goroutine's initial stack is 2–8 KB and grows dynamically. A server with 10,000 concurrent WebSocket connections therefore consumes roughly 20–80 MB of stack memory in Go. The equivalent Node.js implementation using a single event loop with callbacks would avoid parallel stack consumption, but would struggle with CPU-bound tasks (trust score computations, Argon2id hashing) without offloading to Worker Threads, reintroducing concurrency complexity without the type safety of Go's interfaces. Python's GIL fundamentally prevents parallelism on CPU-bound work; even with `asyncio`, Argon2id hashing blocks the event loop unless wrapped in a `run_in_executor` thread pool, adding latency and conceptual overhead.

**Type safety and maintainability.** Clean Architecture as implemented in TrueConnect depends on strict interface contracts — for example, `repository.MatchRepository` defines eight method signatures that the PostgreSQL adapter must satisfy at compile time. Go's structural typing means the compiler verifies this contract on every build, catching interface drift before it reaches staging. TypeScript offers comparable guarantees only when strict mode is enforced throughout a project and type imports are maintained carefully. Python's type hints remain advisory without a CI mypy step, which is error-prone in team settings.

**Binary deployment.** The Go backend compiles to a single static binary (~8 MB) that runs on `alpine:3.19` with no runtime dependency beyond the operating system kernel. The Docker image therefore measures approximately 12 MB. Node.js and Python require their respective runtimes plus all `node_modules` or `site-packages`, producing images typically 15–20× larger. On a Kazakhstan VPS with constrained bandwidth, this materially affects both deployment time and the risk of supply-chain vulnerabilities introduced through transitive npm or pip dependencies.

**Race detection.** The codebase's 102 service-layer tests are executed with Go's built-in `-race` detector on every CI run (`.github/workflows/ci.yml`). The race detector instruments memory accesses at compile time and reports data races at runtime — a critical safety net given the concurrent WebSocket hub where multiple goroutines read and write shared connection maps. No equivalent compile-time race instrumentation exists for Node.js; Python's GIL provides implicit mutual exclusion for CPython but at the cost of true parallelism.

### 6.1.3 Decision

Go 1.25 was selected as the backend language. The decisive factors were: (1) the goroutine model mapping cleanly onto the one-goroutine-per-connection WebSocket architecture; (2) compile-time interface enforcement enabling Clean Architecture contracts; (3) the smallest deployment footprint; and (4) built-in race detection critical for the concurrent trust engine and WebSocket hub.

---

## 6.2 Mobile Framework

### 6.2.1 Candidates Evaluated

Three approaches were considered for the mobile client: **Flutter 3.16+ (Dart 3.2)**, **React Native 0.73 (TypeScript)**, and **native development** (Swift/Kotlin per platform). The primary axis of evaluation was development velocity for a single-developer graduate project while maintaining production-quality UI faithful to the Дала Нұры Islamic design system.

### 6.2.2 Comparative Analysis

**Table 6.2 — Mobile Framework Comparison**

| Criterion | Flutter (Dart) | React Native (TS) | Native Swift+Kotlin |
|-----------|---------------|-------------------|---------------------|
| Rendering engine | Skia/Impeller (own canvas, pixel-perfect) | Native bridge (UIKit / Android Views) | Native UIKit / Jetpack Compose |
| UI consistency iOS/Android | Identical pixel-perfect | Near-identical (some platform inconsistencies) | Different per platform |
| Custom painter / canvas | `CustomPainter` API (full 2D canvas) | `react-native-skia` (external lib, less stable) | CoreGraphics / Canvas |
| Animation | `AnimationController`, `Tween`, `AnimatedBuilder` (60/120 fps, built-in) | `Animated` API, Reanimated 3 (external) | UIKit animation / Compose `animateAsState` |
| State management | Riverpod 2.x (compile-time safe, auto-dispose) | Redux Toolkit / Zustand (convention-based) | SwiftUI @State / Compose remember |
| HTTP client | Dio (interceptors, retry, timeout) | Axios / fetch | URLSession / OkHttp |
| WebSocket | `web_socket_channel` (Dart Streams) | `react-native-websocket` / Socket.IO | URLSessionWebSocketTask / OkHttp WS |
| Secure storage | `flutter_secure_storage` (Keychain/Keystore) | `react-native-keychain` | Keychain / Keystore |
| Code sharing % | ~95% between iOS and Android | ~85% (some native modules needed) | 0% (two codebases) |
| Development effort | 1× (baseline) | 1.2× (bridge debugging) | 2.5× (two codebases) |
| Hot reload | Yes (full state preservation) | Yes (Fast Refresh, occasional state loss) | Limited (Xcode preview / Compose preview) |
| Islamic font support | `google_fonts` (Amiri Arabic, full RTL) | `react-native-localize` + custom fonts | Built-in RTL, custom font loading |
| App bundle size | ~18 MB (release, tree-shaken) | ~22 MB (JS bundle + native modules) | ~8-12 MB |

**Custom rendering for the Дала Нұры design system.** The TrueConnect UI requires three custom-painted Islamic geometric elements: `HalalPatternPainter` (Қошқар мүйіз scrollwork background), `IslamicStarWidget` (8-pointed animated star on the splash screen), and `KazakhDivider` (traditional border motif). Flutter's `CustomPainter` exposes a raw 2D canvas (Skia) with direct access to `Path`, `Paint`, `Canvas.drawPath`, and transformation matrices. These elements were implemented entirely in Dart without any native code or external plugin dependency. The equivalent implementation in React Native would require `react-native-skia`, which is a community library with a separate release cycle and historically lagged behind React Native's own versions, introducing upgrade friction. Native implementations would require separate Swift (CoreGraphics) and Kotlin (Canvas API) versions, doubling maintenance cost.

**Riverpod's compile-time safety.** The nine Riverpod providers in `frontend/lib/features/*/` are typed via Dart's strong type system. `AsyncNotifierProvider<MatchingNotifier, List<Map<String, dynamic>>>` is resolved at compile time; a mismatched `ref.watch()` is a compile error rather than a runtime crash. The auto-dispose lifecycle prevents memory leaks from abandoned subscription streams — particularly valuable for `chatNotifierProvider`, which holds an open `WebSocketChannel` that must be closed when the user navigates away from the chat screen.

**Dart Streams for WebSocket.** The `chatNotifierProvider` in `frontend/lib/features/chat/chat_provider.dart` models the WebSocket as a `Stream<dynamic>`, consuming it via `stream.listen()` with a typed `_handleMessage()` dispatch. Dart's asynchronous Streams are a first-class language construct with structured concurrency semantics; backpressure, cancellation, and error propagation are handled uniformly. React Native's equivalent requires manual event emitter wiring through a JavaScript-to-native bridge, adding latency and error surface. Native iOS `URLSessionWebSocketTask` is well-designed but Kotlin's coroutine-based `OkHttp` WebSocket requires separate implementation.

### 6.2.3 Decision

Flutter 3.16+ with Dart 3.2 was selected. The critical factors were: (1) `CustomPainter` enabling pixel-perfect Islamic geometric art without external dependencies; (2) Riverpod's compile-time type safety for the complex 9-provider state graph; (3) ~95% code sharing reducing a solo developer's workload; and (4) `google_fonts` providing the Amiri Arabic typeface with full Unicode bidirectional text support required for Quranic verse display (Ar-Rum 30:21) on the splash screen.

---

## 6.3 Primary Relational Database

### 6.3.1 Candidates Evaluated

For the primary persistence layer, three databases were evaluated: **PostgreSQL 16 + PostGIS**, **MongoDB 7 (document store)**, and **Firebase Realtime Database / Firestore**. Evaluation criteria were: relational integrity for the 21-table schema, geospatial query capability, encryption at rest, data sovereignty (Kazakhstan law requirement), and cost at scale.

### 6.3.2 Comparative Analysis

**Table 6.3 — Primary Database Comparison**

| Criterion | PostgreSQL 16 + PostGIS | MongoDB 7 | Firebase Firestore |
|-----------|------------------------|-----------|-------------------|
| Data model | Relational (ACID, full SQL) | Document (BSON, flexible schema) | Document (JSON, limited joins) |
| FK enforcement | Native (`FOREIGN KEY`, cascades) | Application-level only | Application-level only |
| Geospatial queries | PostGIS: `ST_DWithin`, `ST_Distance`, `GEOGRAPHY` type | `$geoNear` (2dsphere index) | Geohash-based (GeoPoint, no distance queries) |
| Full-text search | `tsvector` / `pg_trgm` | MongoDB Atlas Search (external) | Requires Algolia / Elasticsearch |
| ACID transactions | Multi-statement, cross-table | Multi-document (4.0+, with overhead) | Limited (single document atomic) |
| Encryption at rest | `pgcrypto`, column-level AES | Field-level encryption (Enterprise) | AES-256 (managed, US jurisdiction) |
| Row-level security | Native (PostgreSQL RLS policies) | Collection-level security rules | Security rules (coarser granularity) |
| Data sovereignty | Self-hosted KZ VPS | Self-hosted or Atlas (region selection) | Google Cloud (US-based, GDPR issues) |
| Schema migrations | `golang-migrate` (versioned SQL files) | Manual or `migrate-mongo` | Schema-less (no migration tooling) |
| Two-schema isolation | Native (`social` + `identity_vault`) | Database-level separation only | No namespace isolation |
| Cost (self-hosted) | Free (open-source) | Free (Community) | Billing per read/write ($0.06/100K reads) |
| Operational complexity | Medium (requires DBA for tuning) | Medium | Low (managed, but no control) |

**PostGIS for proximity matching.** TrueConnect's candidate discovery query uses PostGIS's `ST_DWithin(p.location, ST_MakePoint($lon, $lat)::geography, $radius_m)` to filter candidates within a configurable radius (default 50 km, adjustable via `social.user_settings.max_distance_km`). PostGIS stores locations as the `GEOGRAPHY` type with SRID 4326 (WGS 84), which computes distances correctly on the spherical Earth model. MongoDB's `$geoNear` aggregation stage also supports spherical distance queries on a `2dsphere` index and would be a viable alternative. Firebase's `GeoPoint` type stores latitude/longitude but provides no server-side distance query; implementing proximity search would require a client-side geohash approach or a Cloud Function, adding latency and cost per query.

**Two-schema identity vault.** Kazakhstan's data sovereignty requirements mandate that citizens' IIN (individual identification number) and full name remain on KZ-resident servers. The `identity_vault` schema in PostgreSQL is accessible only to the restricted `identity_vault_role` database role; the application's main connection user (`trueconnect_app`) cannot read vault tables directly. This role-based schema isolation is a native PostgreSQL feature with no equivalent in MongoDB (which provides database-level separation only) or Firestore (which has no concept of schemas).

**ACID transactions for match atomicity.** The matching algorithm must atomically: (1) insert a row into `social.likes`, (2) query whether the reverse like exists, and (3) if so, insert a row into `social.matches` and return the new match ID. This three-step operation is wrapped in a `BEGIN`/`COMMIT` PostgreSQL transaction in `internal/adapter/postgres/match_repo.go`. If any step fails, the entire transaction rolls back — preventing phantom matches or duplicate like records. MongoDB 4.0+ supports multi-document transactions, but with measurable write amplification overhead (the WiredTiger storage engine must track transaction state across documents). Firestore's atomic operations are limited to a single document or a batch of writes with no conditional logic, making the like-then-match pattern awkward to implement without Cloud Functions.

**golang-migrate versioned migrations.** The 19 SQL migration files in `migrations/` (format: `000001_init.up.sql` through `000019_add_cascade_deletes.up.sql`) provide a deterministic, version-controlled schema evolution path. Each migration has a paired `.down.sql` for rollback. MongoDB's schemaless nature eliminates migration files but transfers schema management to application code, which is fragile at scale — a partially deployed application may write documents in new format while old instances still run.

### 6.3.3 Decision

PostgreSQL 16 with the PostGIS extension was selected. The decisive factors were: (1) PostGIS `ST_DWithin` for the location-based candidate filter without external services; (2) native schema isolation (`social` + `identity_vault`) required by Kazakhstan data law; (3) ACID multi-statement transactions for atomic like/match detection; and (4) zero licensing cost with full self-hosting capability on the project VPS.

---

## 6.4 Social Graph Database

### 6.4.1 The Case for a Dedicated Graph Store

Computing trust scores via weighted PageRank and detecting Sybil clusters via Louvain community detection are graph algorithms. While PostgreSQL with recursive CTEs can represent graph traversals, the query complexity grows polynomially with depth; a six-hop relationship chain in PostgreSQL requires a recursive CTE with up to O(n^k) intermediate rows. Neo4j's native graph storage (index-free adjacency) reduces multi-hop traversal to O(k) pointer dereferences regardless of the total graph size.

**Table 6.4 — Social Graph Options**

| Criterion | Neo4j 5 Community + GDS | PostgreSQL (recursive CTE) | Amazon Neptune |
|-----------|------------------------|---------------------------|----------------|
| Storage model | Native graph (index-free adjacency) | Adjacency list in relational tables | Property graph / RDF |
| Multi-hop traversal | O(k) pointer dereferences | O(n^k) recursive CTE | O(k) native (TinkerPop Gremlin) |
| PageRank | GDS library (`gds.pageRank.stream`) | Manual iterative SQL (complex) | Gremlin `PageRankVertexProgram` |
| Louvain clustering | GDS library (`gds.louvain.stream`) | Not available natively | Not available natively |
| Sybil detection | Louvain + degree centrality (built-in) | Requires external Gephi export | Requires external tooling |
| Data sovereignty | Self-hosted KZ VPS | Self-hosted | AWS (US jurisdiction) |
| Cost | Free (Community) | Free (included in PostgreSQL) | ~$0.10/hour minimum |
| Go driver | Official `neo4j-go-driver/v5` | `pgx/v5` | `gremlin-go` |
| Operational complexity | Medium (separate JVM process) | Low (reuse existing DB) | Low (managed) |

**GDS algorithms in production.** The `TrustEngine` worker (`internal/worker/trust_engine.go`) calls `gds.pageRank.stream` with relationship weight set to the product of the rater's verification weight (1.5× for KYC-verified users) and the normalized rating value. This query runs in Neo4j's in-memory projection and returns a stream of (nodeId, score) pairs in milliseconds — a query that would require dozens of iterative SQL passes in PostgreSQL to converge. The Louvain community detection (`gds.louvain.stream`) identifies clusters of users with dense mutual interaction but sparse connections to the rest of the graph, which the Sybil detection logic interprets as fake account rings.

**PostgreSQL as a fallback.** Trust scores are written to both Neo4j (canonical) and `social.users.trust_score` (denormalized copy) to allow SQL joins on trust score during candidate discovery queries without a Neo4j round-trip per row. This write-through pattern combines the query expressiveness of Neo4j with the join performance of PostgreSQL.

### 6.4.2 Decision

Neo4j 5 Community with the Graph Data Science plugin was selected. PostgreSQL alone cannot express PageRank or Louvain efficiently; Neptune satisfies technical requirements but violates Kazakhstan data sovereignty. Neo4j's self-hosted Community edition is free, supports the official Go driver, and the GDS library provides both required algorithms natively.

---

## 6.5 Real-Time Communication Protocol

### 6.5.1 Candidates Evaluated

Real-time bidirectional messaging for TrueConnect's chat and mahram supervision features requires a low-latency, persistent-connection protocol. Three approaches were evaluated: **WebSocket (RFC 6455)**, **Server-Sent Events (SSE) + HTTP POST for client-to-server**, and **long polling**.

**Table 6.5 — Real-Time Protocol Comparison**

| Criterion | WebSocket (RFC 6455) | SSE + HTTP POST | Long Polling |
|-----------|---------------------|-----------------|--------------|
| Connection type | Full-duplex persistent TCP | Half-duplex (server push only) | Request-response loop |
| Latency (median) | 1-5 ms | 10-30 ms (reconnect overhead) | 100-500 ms (poll interval) |
| Server memory per connection | ~4 KB (goroutine stack, Go) | ~2 KB (SSE goroutine) | Negligible (ephemeral handlers) |
| Bidirectional | Yes (client and server send at any time) | No (client POSTs separately) | No (request-response only) |
| Typing indicators | Trivially sent as WS message type | Requires separate POST endpoint | Not practical |
| WebRTC signaling | Directly via `webrtc_offer/answer/ice_candidate` message types | Possible but awkward (POST + SSE receive) | Not practical |
| Nginx proxying | `proxy_pass` + `Upgrade` header + 86400s timeouts | Standard HTTP proxy | Standard HTTP proxy |
| Firewall traversal | Port 80/443 (via HTTP Upgrade) | Port 80/443 | Port 80/443 |
| Mobile reconnect | Exponential backoff (1/2/4/8/16 s) | EventSource auto-reconnect | Application-level retry |
| Flutter library | `web_socket_channel` (official) | `http` package (SSE parsing) | `http` package (polling loop) |

**Full-duplex requirement.** TrueConnect's WebSocket protocol carries eight distinct client-to-server message types (`auth`, `chat_msg`, `typing`, `read`, `webrtc_offer`, `webrtc_answer`, `webrtc_ice_candidate`, `mahram_chat_msg`) and seven server-to-client message types (`auth_ok`, `chat_msg`, `mahram_chat_msg`, `content_warning`, `content_blocked`, `match_notification`, `typing`). SSE would handle server-to-client delivery adequately, but each of the eight client-to-server message types would require a separate HTTP POST endpoint, introducing eight additional API surface points and doubling the network round-trips for interactive operations such as typing indicators (which fire on every keypress).

**WebRTC signaling.** The `CallScreen` uses WebRTC for video calls, requiring an ICE candidate exchange (typically 3–10 candidates per peer) and SDP offer/answer negotiation. The WebSocket hub acts as the signaling channel: `webrtc_offer`, `webrtc_answer`, and `webrtc_ice_candidate` message types are routed to the target user's goroutine via the Hub's `connections` map. This design reuses the existing authenticated WebSocket connection for signaling — no separate signaling server is required. Long polling cannot carry signaling traffic at acceptable latency; SSE + HTTP POST could in principle, but loses the authentication context that is already established on the WebSocket connection.

**Redis Pub/Sub for horizontal scaling.** When a chat message is delivered via POST or WebSocket to one API instance, it must reach the recipient's WebSocket connection even if that connection is held by a different API instance. The Hub publishes to a Redis channel `chat:<matchId>` on every outgoing message; all Hub instances subscribe and deliver to their locally connected users. This fan-out pattern works identically for WebSocket messages. SSE would require the same Redis pub/sub backing, but the separate POST path for client messages would add a second Redis channel type, increasing operational complexity.

**Nginx configuration.** The TrueConnect nginx configuration in `deployments/nginx/nginx.conf` sets `proxy_read_timeout 86400s` and `proxy_send_timeout 86400s` on the `/v1/ws` location block, with `proxy_http_version 1.1`, `proxy_set_header Upgrade $http_upgrade`, and `proxy_set_header Connection "upgrade"`. This is the standard nginx WebSocket proxy configuration. SSE and long polling would use standard HTTP proxy settings but would generate more total requests — SSE because of periodic reconnects (EventSource reconnects on connection loss, defaulting to 3 s), long polling because of the request-response loop.

### 6.5.2 Decision

WebSocket (RFC 6455) via `coder/websocket` was selected. The full-duplex requirement from WebRTC signaling alone justifies the choice; typing indicator frequency (up to 10 events per second per user) over SSE would require 10 POST requests per second per active typist; and the authentication context established at WebSocket open eliminates per-request JWT verification overhead for the high-frequency message types.

---

## 6.6 Caching and Session Store

Redis 7 was selected over Memcached and over storing sessions in PostgreSQL based on three functional requirements specific to TrueConnect:

1. **WebSocket Pub/Sub.** Redis's native Pub/Sub is used to fan out chat messages across Hub instances. Memcached has no pub/sub primitive; PostgreSQL `LISTEN`/`NOTIFY` is limited to 8 KB payloads and does not scale past a few hundred concurrent listeners.

2. **Sliding-window rate limiting.** Per-user rate limiting uses Redis sorted sets (`ZADD` + `ZREMRANGEBYSCORE` + `ZCARD`) implementing the sliding window algorithm. The atomic Lua script evaluates the window and increments the counter in a single round-trip. PostgreSQL could implement the same with a `rate_limit_events` table, but would require a row-level lock per request — unacceptable for the authentication endpoints limited to 5 req/s.

3. **Candidate seen-set TTL.** The matching algorithm tracks which candidates a user has already seen using a Redis set keyed `seen:<userId>` with a 24-hour TTL. Redis's native key expiration avoids a background cleanup job. PostgreSQL's `social.swipe_rejections` table stores permanent pass decisions, but the ephemeral 24-hour seen-set (which excludes previously shown profiles from the current session's batch) is unsuited to a relational table.

**Table 6.6 — Cache Store Comparison**

| Criterion | Redis 7 | Memcached 1.6 | PostgreSQL (session table) |
|-----------|---------|---------------|---------------------------|
| Pub/Sub | Native (`PUBLISH`/`SUBSCRIBE`) | Not available | `LISTEN`/`NOTIFY` (8KB limit) |
| Key TTL | Per-key configurable | Per-item TTL | Requires background job |
| Sorted sets | Native (`ZADD`, `ZRANGE`) | Not available | Requires table + index |
| Lua scripting | `EVAL` atomic scripts | Not available | PL/pgSQL functions |
| Persistence | RDB snapshots + AOF | None | Full ACID |
| Max memory policy | `allkeys-lru` (configured) | `allkeys-lru` | N/A |
| Data structures | Strings, sets, sorted sets, hashes, streams | Strings only | Full SQL types |

Redis 7 is configured with `maxmemory 256mb` and `maxmemory-policy allkeys-lru` in the Docker Compose file. This ensures that under memory pressure, the least recently used cached entries (primarily candidate seen-sets and non-critical caches) are evicted before critical operational data (session tokens use explicit TTLs and are replicated to PostgreSQL).

---

## 6.7 Object Storage

MinIO was selected over AWS S3 and Google Cloud Storage for a single decisive reason: **data sovereignty**. Kazakhstan's Law on Personal Data (2013, amended 2023) requires that personal data of Kazakhstan citizens be stored within the Republic of Kazakhstan. MinIO runs as a self-hosted S3-compatible object store on the same KZ-resident VPS as the rest of the stack, ensuring that avatar photos and KYC identity documents never leave Kazakhstan jurisdiction. The S3 API compatibility means the `minio-go/v7` client SDK is API-identical to the AWS SDK, enabling a future migration to AWS S3 (for non-PII media) without code changes if the legal landscape evolves.

---

## 6.8 Summary

**Table 6.7 — Technology Selection Summary**

| Layer | Selected | Primary Justification |
|-------|----------|-----------------------|
| Backend language | Go 1.25 | Goroutine-per-connection WS model; compile-time interface enforcement; `-race` detector |
| Mobile framework | Flutter 3.16 (Dart 3.2) | `CustomPainter` for Islamic geometric art; Riverpod compile-time safety; 95% code sharing |
| Relational database | PostgreSQL 16 + PostGIS | PostGIS proximity queries; dual-schema identity vault; ACID match atomicity |
| Graph database | Neo4j 5 + GDS | GDS PageRank and Louvain; O(k) traversal; self-hosted for KZ data sovereignty |
| Real-time protocol | WebSocket (RFC 6455) | Full-duplex for WebRTC signaling; typing indicators; Redis pub/sub fan-out |
| Cache / session store | Redis 7 | Pub/Sub; sliding-window rate limiting; seen-set TTL eviction |
| Object storage | MinIO | KZ data sovereignty for photos and KYC documents; S3-compatible API |

Each selection is grounded in a specific, verifiable technical requirement of the TrueConnect system rather than general popularity. The combination of Go's concurrency model with Flutter's rendering engine, PostgreSQL's relational integrity with Neo4j's graph algorithms, and WebSocket's full-duplex channel with Redis's pub/sub creates a coherent technology stack in which each component addresses limitations of its alternatives that would otherwise require architectural workarounds or additional infrastructure.
