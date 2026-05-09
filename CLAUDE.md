# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TrueConnect is a trust-based social networking and dating platform for Kazakhstan. It consists of a Go backend (modular monolith), a Flutter mobile app, and a Next.js web app. The defining feature is a reputation/trust engine backed by a Neo4j graph database that computes weighted trust scores and detects Sybil attacks.

## Development Commands

### Backend (Go — root of repo)
```bash
make run            # Start the API server locally
make run-worker     # Start the background worker locally
make test           # Run all tests (with race detector)
make lint           # Run golangci-lint
make build          # Compile api, worker, and migrate binaries to bin/
make migrate-up     # Apply SQL migrations
make migrate-down   # Rollback last migration

make docker-up      # Build and start all services via docker-compose
make docker-down    # Stop all services
make docker-reset   # Tear down volumes + rebuild (full reset)
make docker-logs    # Tail all service logs
```

Run a single test file or package:
```bash
go test ./internal/service/... -v -run TestAuthService
```

### Web (Next.js — `web/`)
```bash
npm run dev    # Development server
npm run build  # Production build
npm run lint   # ESLint
```

### Mobile (Flutter — `frontend/`)
```bash
flutter run                  # Run on connected device
flutter run -d chrome        # Run in browser
flutter run -d android
```

## Environment Setup

Copy `.env.example` to `.env`. Key variables:
- `ENCRYPTION_KEY` — 32-byte AES key, hex-encoded. Generate: `openssl rand -hex 32`
- `JWT_SECRET` — must be ≥32 characters
- `FIREBASE_CREDENTIALS` — path to Firebase service account JSON (optional; falls back to mock push provider)

Full stack runs via `deployments/docker-compose.yml` (PostgreSQL, Neo4j, Redis, MinIO, Nginx + the Go API).

## Backend Architecture

The backend is a **strict clean architecture** monolith. Dependencies flow in one direction only:

```
handler → service → repository (interface) ← adapter (implementation)
                 ↑
              domain (entities, no external deps)
```

- `internal/domain/` — pure Go structs and domain errors; no imports from the project
- `internal/repository/` — interface definitions (ports); services depend only on these
- `internal/adapter/` — concrete implementations: `postgres/`, `neo4j/`, `redis/`, `minio/`, `kyc/`
- `internal/service/` — business logic; receives repository interfaces via constructor injection
- `internal/handler/` — Gin HTTP handlers, middleware, WebSocket hub; calls services
- `internal/pkg/` — shared utilities (`crypto`, `jwt`, `logger`, `sanitize`, `validator`)
- `internal/worker/` — background goroutines: `TrustEngine` (reputation), `PushWorker` (FCM)
- `cmd/api/` — wires everything together via `main.go`; `cmd/worker/` and `cmd/migrate/` are separate entry points

**Critical rule**: services must never import adapter packages directly — only repository interfaces.

## Data Layer

| Store | Role |
|---|---|
| PostgreSQL | Primary relational store: users, profiles, posts, messages, matches, notifications, KYC vault |
| Neo4j | Trust graph: `(:User)-[:RATED]->(:User)`, `[:INTERACTED_WITH]`, Sybil detection clusters |
| Redis | Sessions, rate limiting, matching candidate cache, reputation score cache, WebSocket Pub/Sub |
| MinIO | Object storage: avatars, photos, KYC documents |

**Dual-schema identity vault**: PostgreSQL uses two schemas — `social` (public data) and `identity_vault` (encrypted PII: IIN, full name, documents). The vault schema is only accessible via a restricted DB role. Every vault access is audit-logged.

## Key Systems

### Authentication
- JWT HS256: 15-minute access tokens, 7-day refresh tokens (one-time-use rotation)
- Refresh tokens are hashed (Argon2id) before storage in PostgreSQL; the raw token is returned once
- Passwords: Argon2id (64 MB memory, 3 iterations, 4 parallelism)
- PII fields: AES-256-GCM with a 12-byte random nonce per field

### Reputation / Trust Engine (`internal/worker/trust_engine.go`, `internal/service/reputation_service.go`)
Scores are computed asynchronously in Neo4j when an interaction is confirmed. Formula:
1. Each rating is weighted by the rater's identity verification (1.5× if KYC verified) and their own trust score (rater_score / 100)
2. Bayesian smoothing toward a 2.5/5 neutral baseline (C=5 virtual ratings)
3. Final score = smoothed × 20 → integer 0–100

Scores are written to Neo4j, PostgreSQL (for SQL joins), and Redis (sub-millisecond swiping feed).

### Sybil Detection
Neo4j Louvain community detection runs every 6 hours (via cron in the worker). Clusters with <30% external edges and low KYC verification rates are flagged.

### Location-Based Matching
PostGIS `ST_DWithin` queries on the `social` schema. Matching candidates are cached in Redis with configurable TTL.

### Real-Time Chat
WebSocket hub (`internal/handler/chat_handler.go`). Messages are encrypted with AES-256-GCM before storage. Redis Pub/Sub fans out messages across multiple API instances.

## Frontend Architecture

**Flutter (`frontend/`)**: Riverpod for state management, GoRouter for navigation, Dio for HTTP, `web_socket_channel` for WebSocket, `flutter_secure_storage` for tokens, `json_serializable` + `build_runner` for model codegen.

**Next.js (`web/`)**: TanStack React Query for data fetching, Zustand for global state, Tailwind CSS, TypeScript, `react-hook-form`, `recharts` for analytics.

## Migrations

SQL migrations live under `migrations/` (golang-migrate format). Each migration requires a paired `up` and `down` file. Run `make migrate-up` / `make migrate-down` or use the `cmd/migrate` binary directly.

## Reference Docs

- `docs/ARCHITECTURE.md` — full system architecture, DDL schemas, API contract, development roadmap
- `docs/REPUTATION_ENGINE.md` — detailed trust score formula and Bayesian smoothing explanation
- `docs/MEMORY.md` — project memory / decision log
- `api/openapi.yaml` — OpenAPI specification
