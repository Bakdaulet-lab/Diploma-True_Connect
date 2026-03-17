# TrueConnect Project Memory

## Project Overview
- **Name:** TrueConnect — trust-based social networking/dating platform for Kazakhstan
- **Domain:** Online dating with identity verification & reputation scoring
- **Architecture doc:** `docs/ARCHITECTURE.md` (v1.0, created 2026-03-16)

## Tech Stack
- **Backend:** Go (Gin), clean architecture monorepo, modular monolith
- **Web:** Next.js (SSR)
- **Mobile:** Flutter
- **DB:** PostgreSQL (PostGIS) + Neo4j (trust graph) + Redis (cache/sessions)
- **Storage:** MinIO (S3-compatible, on KZ VPS)
- **Infra:** Docker Compose, Nginx reverse proxy, deployed on KZ VPS

## Key Architecture Decisions
- Monorepo with `cmd/`, `internal/` (domain, repository, service, adapter, handler, worker, pkg)
- Interface-driven design: repositories are interfaces, adapters implement them
- Identity vault: separate PG schema (`identity_vault`) with restricted access, AES-256 encrypted PII
- JWT (15min) + HttpOnly refresh tokens (7 days, one-time-use rotation)
- Trust score: 0-100, computed via weighted PageRank on Neo4j graph, cached in Redis
- Sybil detection: Neo4j GDS Louvain community detection, cron every 6 hours
- All PII on KZ-local servers only (data sovereignty)

## Key Packages
- pgx v5, sqlc, golang-migrate, neo4j-go-driver v5, go-redis v9
- minio-go v7, coder/websocket, go-playground/validator v10, golang-jwt v5
- Argon2id (golang.org/x/crypto/argon2), AES-256-GCM for encryption at rest

## Sprint Status
- Sprint 0 (scaffolding): NOT STARTED
- Sprint 1 (auth): NOT STARTED
- Sprint 2 (profiles/matching): NOT STARTED
- Sprint 3 (reputation/graph): NOT STARTED
- Sprint 4 (chat/feed/polish): NOT STARTED

## Coding Rules (for future code generation)
- See `docs/ARCHITECTURE.md` Section 5 for full rules
- Never silence errors; wrap with fmt.Errorf + %w
- Context propagation on all I/O functions
- Parameterized queries only (no string concat SQL)
- No global state; constructor injection
- Honesty rule: never invent fake packages or APIs

## User Profile
- Software Engineering student, graduation project / startup MVP
- I act as Tech Lead providing production-grade guidance
