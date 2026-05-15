# Chapter 4: Methodology

## 4.1 Development Methodology: Agile with Scrum

TrueConnect was developed using an Agile methodology structured around iterative sprints, each delivering a vertical slice of production-quality functionality. This choice was motivated by the specific nature of the problem domain: Islamic matrimony application design is a novel engineering challenge with few established precedents, meaning that requirements could not be fully specified in advance and needed to emerge through iterative development and continuous validation against domain knowledge.

The project was organised into fourteen sprints (Sprints 0–13), each focused on a coherent functional area. Sprint 0 established the foundational scaffolding: the monorepo structure, Docker Compose orchestration, database connection pool initialisation, health check endpoints, and all inter-service connectivity. Sprint 1 delivered complete authentication infrastructure (registration, login, JWT issuance, Argon2id password hashing, AES-256-GCM PII field encryption, refresh token rotation). Sprints 2 through 5 progressively built the core matchmaking features (profiles, discovery, matching, settings, photo upload, social feed, KYC stub, WebSocket chat, input sanitisation). Sprints 6 through 9 delivered the Islamic-domain-specific features that differentiate TrueConnect from all competitors: the halal identity schema extensions (`migrations/000013_halal_identity_fields.up.sql`, `000014_mahram.up.sql`, `000015_halal_matches_ext.up.sql`), the niyyah compatibility filter (B1), the madhab affinity boost (B2), the no-photo modesty enforcement (B3), and the niyyah 90-day timer. Sprints 10 through 11 completed the mahram chat architecture, the family introduction milestone, the embedded imam catalog, and the admin whisper-flag review system. Sprints 12 and 13 delivered the complete Flutter mobile frontend across 21 screens with the Дала Нұры design system.

User stories throughout the project were expressed in Islamic domain language, which directly informed the ubiquitous language of the codebase. Representative examples include: "As a Muslim woman seeking marriage, I want my photos hidden from non-mahram men until I choose to share them, so that my digital presence respects Islamic modesty norms" — which produced the `no_photo_mode` architectural feature; "As a user with nikah_year niyyah, I want to only see candidates who share my serious marital intent, so that my time is not wasted on incompatible interactions" — which produced the `AllowedNiyyahs` filter in `FindCandidatesOpts`; and "As a Muslim woman, I want my mahram to be an active participant in my conversations, not just notified after the fact" — which produced the mahram chat room system.

Backlog prioritisation followed a dependency-driven sequence: authentication was required before any feature could be built; profile creation was required before matching; matching was required before chat; and Islamic domain extensions were layered onto the stable core in later sprints. Retrospective learning shaped the architecture: early sprints revealed that Islamic features added as overlays to a secular core were fragile and required workarounds, which led to the decision in Sprint 6 to add Islamic fields at the schema level rather than as application-layer annotations. This architectural insight — that Islamic compliance must be embedded in the data model, not applied to it — is the single most important methodological lesson of the project.

---

## 4.2 Domain-Driven Design Application

Domain-Driven Design (DDD) provided the structural vocabulary that allowed TrueConnect's complex Islamic domain to be expressed faithfully in code. The application of DDD to this project is not merely theoretical: the bounded contexts, ubiquitous language terms, aggregates, and repository abstractions described below are directly observable in the Go source code, PostgreSQL schema, and Flutter model definitions.

### 4.2.1 Identifying Bounded Contexts

DDD prescribes that a complex domain be partitioned into bounded contexts — sub-domains with their own consistent model and clear boundaries. Each TrueConnect backend module corresponds to a bounded context with well-defined responsibilities and interface contracts:

**Auth Context** (`internal/handler/auth_handler.go`, `internal/service/auth_service.go`): Responsible for identity establishment and session lifecycle. Owns the concepts of credential, token, and session. Communicates with other contexts only through the `User.ID` (a UUID that serves as the cross-context identity key) and the `VerificationLevel` enum that signals KYC status to the Reputation context.

**Profile Context** (`internal/handler/profile_handler.go`, `internal/service/profile_service.go`): Responsible for the user's public presentation and Islamic identity attributes. Owns `Profile`, `PromptAnswer`, `Niyyah`, `Madhab`, and `Media`. The profile context is the canonical owner of the Islamic domain vocabulary — the fields `niyyah`, `madhab`, `no_photo_mode`, and `languages` are defined here and consumed by the Discovery context's filters.

**Discovery Context** (`internal/handler/matching_handler.go`, `internal/service/matching_service.go`): Responsible for the candidate recommendation algorithm and the like/pass interaction. Owns `LikeResult`, `CandidateCard`, and `FindCandidatesOpts`. Consumes profile data from the Profile context and settings data from the Settings context. The niyyah and madhab filtering logic — the algorithmic expression of Islamic compatibility — lives entirely within this bounded context.

**Chat Context** (`internal/handler/chat_handler.go`, `internal/service/chat_service.go`): Responsible for real-time messaging and the WebSocket Hub. Owns `Message`, `Hub`, and `Client`. Messages are encrypted before storage and decrypted on retrieval entirely within this context, ensuring that no other context ever handles plaintext message content.

**Mahram Context** (`internal/handler/mahram_handler.go`, `internal/service/mahram_service.go`): Responsible for the guardian supervision infrastructure. Owns `Mahram`, `MahramRoom`, and `MahramMessage`. This context is architecturally isolated because its access control rules — who can read and write messages in a mahram channel — are specific and cannot be generalised to the Chat context without conflating the two distinct communication models.

**Reputation Context** (`internal/service/interaction_service.go`, `internal/service/reputation_service.go`, `internal/worker/trust_engine.go`): Responsible for the community trust score system. Owns `Interaction`, `TrustScore`, and `LeaderboardEntry`. Communicates with the Neo4j adapter for graph-based score computation and with Redis for score caching.

**Feed Context** (`internal/handler/post_handler.go`, `internal/service/post_service.go`): Responsible for the community social feed. Owns `Post`, `PostComment`, and like/unlike operations. The reputation gate in `CreatePost` — which requires a minimum trust score of 30 — is a dependency on the Reputation context's score data.

**Settings Context**, **KYC Context**, **Notifications Context**, **Whisper Context**, **Imam Context**, and **Admin Context** each constitute smaller but equally well-bounded contexts owning their respective domain concepts.

Context boundaries are enforced at the Go package level: the `internal/service/auth_service.go` file imports only `internal/domain/` and `internal/repository/` interfaces — never `internal/adapter/` packages — ensuring that the business logic of one context cannot directly manipulate the data stores of another.

### 4.2.2 Ubiquitous Language

The DDD concept of ubiquitous language — identical terminology used by domain experts, developers, designers, and the codebase — is demonstrably implemented in TrueConnect. The following Islamic domain terms appear with identical spelling and semantics across all system layers:

**Niyyah (نية)**: Appears as `domain.NiyyahNikahYear`, `domain.NiyyahSeriousMarriage`, `domain.NiyyahFriendship` (Go enums in `internal/domain/profile.go`); as the `niyyah` column in `social.profiles` (PostgreSQL schema); as `AllowedNiyyahs []string` in `repository.FindCandidatesOpts`; as the `/niyyah` route in GoRouter (`lib/core/router/app_router.dart`); as `Profile.niyyah` in the Flutter model (`lib/models/profile.dart`); and as the `NiyyahBadge` widget (`lib/widgets/niyyah_badge.dart`). The term travels from the Islamic jurisprudential concept through the UX layer without any semantic transformation.

**Mahram (محرم)**: Appears as `domain.Mahram` (Go struct); as `social.mahrams` and `social.mahram_chat_rooms` (PostgreSQL tables); as the `POST /v1/mahram` API endpoint; as the `mahram_chat_msg` WebSocket message type; as the `/mahram-chat/:roomId` route in GoRouter; and as `MahramChatScreen` in Flutter. The term's technical implementation — a third-party participant in an encrypted three-way communication channel — is a precise digital translation of the Islamic guardianship concept.

**Madhab (مذهب)**: Appears as `domain.MadhabHanafi`, `domain.MadhabShafii`, `domain.MadhabMaliki`, `domain.MadhabHanbali`, `domain.MadhabNone` (Go enums); as the `madhab` column in `social.profiles`; as `Profile.madhab` in Flutter; and as the `MadhabBadge` widget. The madhab affinity boost in `matching_service.go` (B2 feature) is the algorithmic expression of this term's significance.

**Match**: Appears as `domain.Match` (Go struct with `UserAID`, `UserBID`, `MatchedAt`); as `social.matches` (PostgreSQL table); as `GET /v1/matches` (API endpoint); as `Match` (Flutter model with computed `daysLeftOnTimer`); and as `MatchesScreen` (Flutter screen). The concept of a bilateral mutual consent — two likes creating a match — is consistent across all layers.

**Like/Pass**: Appears as `matchRepo.RecordLike()` and `matchRepo.RecordPass()`; as `social.likes` and `social.swipe_rejections` tables; as `POST /v1/matching/like` and `POST /v1/matching/pass` endpoints; and as `matchingNotifier.like(id)` and `matchingNotifier.pass(id)` in the Flutter provider. The terminology is invariant across the entire stack.

### 4.2.3 Aggregates and Entities

DDD aggregates are clusters of domain objects that should be treated as a single unit for data consistency purposes, with a root entity that controls access to all members.

**User Aggregate**: Root entity is `domain.User` (identified by `ID uuid`). The aggregate owns `domain.Profile` (one-to-one, always loaded with the user for profile operations), `UserSettings` (discovery preferences), and `RefreshToken` (session lifecycle). The consistency boundary ensures that a user's profile and settings are always associated with a valid, active user account. Deletion of the root (`DELETE /v1/users/me`) cascades to all owned entities within the aggregate.

**Match Aggregate**: Root entity is `domain.Match` (identified by `ID uuid`, referencing `UserAID` and `UserBID`). The aggregate owns the conversation: all `Message` entities in `social.messages` where `match_id` corresponds to this match. The aggregate also owns the milestone progression state: `family_intro_done` and `imam_confirmed` boolean flags. The consistency boundary prevents messages from existing for a non-existent match (enforced by the `match_id` foreign key with ON DELETE CASCADE in migration 000019).

**Post Aggregate**: Root entity is `domain.Post`. The aggregate owns `PostComment` entities and the `post_likes` join table. The like count and comment count are denormalised into the root entity for read performance (incremented/decremented atomically in the repository rather than computed via COUNT queries).

**MahramRoom Aggregate**: Root entity is `MahramRoom` (identified by `ID uuid`, referencing `MatchID` and `MahramUserID`). The aggregate owns `MahramMessage` entities. The three-party access control rule — only the woman, the man from the associated match, and the specified mahram user may read or write messages — is enforced by the `mahram_service.go` layer before any message operation.

### 4.2.4 Repository Pattern

The repository pattern, as applied in TrueConnect, provides a collection-like interface to domain objects while completely hiding the underlying storage technology from the service layer. Each domain aggregate has a corresponding repository interface defined in `internal/repository/`:

```
ProfileRepository     — Upsert, GetByUserID, FindCandidates, GetLeaderboard
MatchRepository       — RecordLike, RecordPass, ListMatches, GetMatch, IsMatched,
                        Unmatch, UnmatchByUsers, BlockUser, GetBlockedIDs, GetRejectedIDs,
                        MarkFamilyIntroDone, MarkImamConfirmed, ListMatchViews,
                        GetPendingLikes, FindExpiredNiyyahMatches
MessageRepository     — Create, ListByMatch, MarkRead
PostRepository        — Create, List, GetByID, LikePost, UnlikePost, AddComment, ListComments
MahramRepository      — CreateMahram, ListMahrams, CreateRoom, ListRooms, ListMessages, SendMessage
SettingsRepository    — Get, Upsert
NotificationRepository — Create, List, MarkRead, MarkAllRead
InteractionRepository — Create, ListByUser, GetCooldown
TrustGraphRepository  — UpsertUser, RecordRating, ComputeScore, DeleteUserNode, RunSybilDetection
```

The concrete implementations of these interfaces live in `internal/adapter/postgres/`, `internal/adapter/neo4j/`, and `internal/adapter/redis/`. Services are constructed by injecting the interface — never the concrete adapter — which provides two critical benefits: full testability (102+ service tests use in-memory mock implementations rather than a real database) and full storage-technology substitutability (the PostgreSQL adapter for `ProfileRepository` could be replaced with a different database without changing any service code).

---

## 4.3 Requirements Engineering

### Functional Requirements

| ID | Requirement | Priority | Source |
|----|-------------|----------|--------|
| FR-01 | The system shall allow users to register with a phone number and password | High | Auth module |
| FR-02 | The system shall issue JWT access tokens (15-minute expiry) and HttpOnly refresh tokens (7-day, one-time-use rotation) | High | Auth module |
| FR-03 | The system shall allow users to declare their niyyah (nikah_year, serious_marriage, friendship) during onboarding | High | NiyyahSelectionScreen |
| FR-04 | The system shall filter discovery candidates to exclude niyyah-incompatible users | High | Discovery module, B1 |
| FR-05 | The system shall allow users to specify their madhab and apply a +10 affinity boost to same-madhab candidates | Medium | Discovery module, B2 |
| FR-06 | The system shall allow users to enable no-photo mode, causing their avatar to be suppressed from all discovery results | High | Profile module, B3 |
| FR-07 | The system shall support mutual-like matching, creating a match record only when both parties have liked each other | High | Matching module |
| FR-08 | The system shall provide real-time bidirectional chat between matched users via WebSocket | High | Chat module |
| FR-09 | The system shall encrypt all chat messages with AES-256-GCM before storage | High | Chat module |
| FR-10 | The system shall allow a woman to add a mahram guardian and create a three-party supervised chat room | High | Mahram module |
| FR-11 | The system shall allow users to submit post-meeting interaction ratings (1–5 stars, context: date/meetup/event) | High | Reputation module |
| FR-12 | The system shall compute a trust score (0–100) for each user using Bayesian-smoothed Neo4j PageRank | High | Trust engine worker |
| FR-13 | The system shall allow users to submit KYC identity documents for verification | High | KYC module |
| FR-14 | The system shall store KYC identity data (IIN, full name) in an AES-256-GCM encrypted identity vault schema | High | Identity vault |
| FR-15 | The system shall provide a community social feed supporting post creation, likes, and comments | Medium | Feed module |
| FR-16 | The system shall detect and flag Sybil attack clusters using Neo4j GDS Louvain community detection every 6 hours | High | Trust engine worker |
| FR-17 | The system shall provide an embedded directory of imams in 5 Kazakhstani cities | Medium | Imam module |
| FR-18 | The system shall record the family introduction milestone and imam nikah confirmation as verifiable match milestones | Medium | Matching module |
| FR-19 | The system shall enforce a 90-day niyyah timer on nikah_year matches, dissolving the match on expiry | Medium | NiyyahTimerWorker |
| FR-20 | The system shall deliver FCM push notifications for new likes and mutual matches | Medium | PushWorker |
| FR-21 | The system shall provide an anonymous whisper reporting system with 3-strike pattern detection | Medium | Whisper module |
| FR-22 | The system shall provide admin endpoints for KYC verdict, Sybil cluster review, and whisper flag review | High | Admin module |

### Non-Functional Requirements

| ID | Requirement | Metric | Implementation |
|----|-------------|--------|----------------|
| NFR-01 | API response time | < 500ms p95 for all REST endpoints | Go + Gin + PostgreSQL with PostGIS indexes |
| NFR-02 | Real-time message delivery latency | < 200ms end-to-end under normal load | WebSocket Hub + Redis pub/sub |
| NFR-03 | Authentication security | JWT HS256, Argon2id (64MB, 3 iter, 4 threads), AES-256-GCM | `internal/pkg/crypto/`, `internal/pkg/jwt/` |
| NFR-04 | Data sovereignty | All user data stored on Kazakhstani VPS | Deployments configuration, MinIO self-hosted |
| NFR-05 | Horizontal scalability | Multiple API instances with shared state | Redis pub/sub WebSocket fan-out, stateless JWT |
| NFR-06 | Availability | > 99% uptime for API service | Docker restart: unless-stopped, health checks |
| NFR-07 | Test coverage | 100% of service-layer business logic covered | 102+ tests, `go test -race`, service_test package |
| NFR-08 | Rate limiting | Auth: 5 req/s; General API: 30 req/s per client | nginx rate limiting zones |
| NFR-09 | File upload security | MIME type validation (JPEG, PNG, PDF only); max 10MB | KYC handler MIME check, nginx `client_max_body_size` |
| NFR-10 | Privacy compliance | KZ Law on Personal Data (2013, amended 2023) | identity_vault restricted PG role, AES-256-GCM |
| NFR-11 | Mobile performance | 60fps swipe animations, < 1s cold start | Flutter Skia/Impeller rendering, Riverpod lazy loading |
| NFR-12 | Offline resilience | Session restore from `flutter_secure_storage` on cold start | AuthNotifier._restoreSession() with background refresh |

---

## 4.4 Testing Strategy

TrueConnect's testing strategy is focused on the service layer — where all Islamic domain business logic resides — and is structured to provide maximum confidence in domain invariant correctness with minimum dependency on external infrastructure.

**Unit Tests (Service Layer):** All 102+ tests reside in `internal/service/` as a separate `service_test` package (black-box testing of exported service interfaces). Tests use in-memory mock implementations of all repository interfaces defined in `internal/service/sprint2_mocks_test.go` and related mock files. Mock implementations precisely replicate the behaviour specified by each repository interface (including error conditions such as `domain.ErrNotFound` and `domain.ErrForbidden`) without requiring a live database connection. All tests are run with Go's `-race` detector enabled (`go test ./internal/service/... -v -race -count=1`) to surface any concurrency bugs in service logic, which is particularly important for the trust score computation goroutines and WebSocket hub operations.

**Key Test Scenarios:** The test suite covers: authentication — token generation and refresh token rotation; profile — niyyah filter exclusion (B1), no-photo mode blur flag (B3); matching — one-sided like produces no match, mutual like produces exactly one match, self-like returns `domain.ErrInvalidInput`, block user removes match; trust score — Bayesian smoothing formula accuracy, KYC weight multiplier, Sybil cluster detection threshold; mahram — third-party validation (mahram cannot be a match participant); settings — default values applied when no settings row exists, pagination bounds clamping.

**Integration Testing:** API endpoint integration tests validate the HTTP layer independently of the database. The CI pipeline (`ci.yml`) runs `go build ./cmd/api` and `go build ./cmd/worker` on every push to confirm compilability across all modules, and `go vet ./...` for static analysis.

**Manual Testing:** WebSocket flows (real-time chat, mahram chat, typing indicators, read receipts) and the full Flutter UI (all 21 screens) were validated through manual testing sessions using the `make seed-halal` demo scenario (Айгерим + Алихан demo users). The complete demo pathway — `make docker-up && make seed-halal` followed by Flutter launch — was validated end-to-end.

**Areas Without Automated Coverage:** The Flutter frontend has no automated test suite in the current version (widget tests and integration tests are planned for a post-graduation sprint). The Neo4j Louvain Sybil detection algorithm is tested indirectly through integration with the live Neo4j instance in the demo environment, not through unit tests.

---

## 4.5 Justification of Methodology

The combination of Agile iteration and Domain-Driven Design was uniquely well-suited to the specific challenges of developing TrueConnect, and the justification for this combination is grounded in both the general software engineering literature and the specific circumstances of this project.

**Why Agile was necessary.** Islamic matrimony application design is a genuinely novel engineering problem. When the project began, there was no reference architecture for a halal-compliant mobile matrimony platform with technical mahram supervision, a community trust score, and a structured Islamic courtship pathway. In this context, a Waterfall methodology — which requires complete requirements specification before implementation begins — would have produced an incorrect specification. The iterative sprint structure allowed domain requirements to emerge progressively: it was only after implementing the basic chat system in Sprint 4 that the specific technical inadequacy of a notification-based "chaperone" feature (as implemented by Muzz and Hawaya) became fully clear, leading to the architectural decision to build a genuine three-party communication channel in Sprint 10. This kind of discovery is only possible with an Agile process.

**Why DDD was necessary.** The Islamic domain is semantically rich and culturally specific. Without DDD's concept of ubiquitous language, the Islamic terms (niyyah, mahram, madhab) would likely have been translated into generic software concepts (intention_type, guardian_id, school_of_thought) — losing the semantic precision that makes the codebase understandable to both Islamic domain experts and software engineers. DDD's bounded context model also provided the architectural justification for isolating the Mahram context from the Chat context: these are distinct Islamic concepts with distinct access control rules, and conflating them would have produced an unmaintainable service with mixed concerns.

**Why the modular monolith was preferable to microservices.** As established in the literature review [19, 20], microservices introduce distributed systems complexity that is only justified at a scale and team size that TrueConnect has not yet reached. The modular monolith provides the module isolation of microservices — each module has a defined interface contract — without the operational overhead of service discovery, distributed tracing, and inter-service network calls. The architecture is designed to support future module extraction (each module already has the interface structure needed to become an independent service) without paying the microservices tax during the initial development phase.

**Why the combined approach produces superior results.** Agile provides the iteration cadence to discover requirements progressively; DDD provides the structural vocabulary to encode those requirements faithfully in code; and the modular monolith provides the deployment simplicity appropriate to the current scale. Neither methodology alone would have been sufficient: DDD without Agile would have produced an over-engineered domain model for requirements that had not yet been discovered; Agile without DDD would have produced a system where Islamic terms were inconsistently named across layers, making the codebase opaque to domain experts. The combination produced a system where the code, the database, the API, and the UI all speak the same Islamic language — which is the central architectural achievement of TrueConnect.
