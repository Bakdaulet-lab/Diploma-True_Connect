# TrueConnect Backend

Trust-based social networking and dating platform for Kazakhstan. Built with Go, PostgreSQL, Neo4j, Redis, and MinIO.

## Tech Stack

- **Runtime:** Go 1.22+
- **Web Framework:** Gin
- **Databases:** PostgreSQL (PostGIS) + Neo4j (trust graph) + Redis (cache/sessions)
- **Object Storage:** MinIO (S3-compatible)
- **Authentication:** JWT + Argon2id + AES-256-GCM encryption

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)

### Using Docker Compose

```bash
# Copy environment template
cp .env.example .env

# Edit .env and set your secrets (especially ENCRYPTION_KEY and JWT_SECRET)

# Start all services (postgres, neo4j, redis, minio, api, worker, migrations)
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down

# Reset everything (removes volumes)
make docker-reset
```

The API will be available at `http://localhost:8080`.

### Local Development

```bash
# Install dependencies
go mod download

# Start infrastructure (postgres, neo4j, redis, minio)
docker compose -f deployments/docker-compose.yml up -d postgres neo4j redis minio

# Run migrations
make migrate-up

# Run the API server
make run

# Run the background worker
make run-worker

# Run tests
make test
```

## Project Structure

```
.
├── api/                    # OpenAPI specification
├── cmd/
│   ├── api/               # HTTP server entrypoint
│   ├── worker/            # Background worker entrypoint
│   └── migrate/           # Database migration tool
├── deployments/           # Docker Compose and deployment configs
├── docs/                  # Architecture documentation
├── internal/
│   ├── adapter/           # Database adapters (postgres, neo4j, redis, minio)
│   ├── config/            # Configuration loading
│   ├── domain/            # Domain models and errors
│   ├── handler/           # HTTP handlers and middleware
│   ├── pkg/               # Shared utilities (jwt, validator, logger, crypto)
│   ├── repository/        # Repository interfaces
│   ├── service/           # Business logic layer
│   └── worker/            # Background job processors
├── migrations/            # SQL migration files
└── scripts/               # Utility scripts
```

## API Endpoints

### Authentication
- `POST /v1/auth/register` - Create new account
- `POST /v1/auth/login` - Authenticate user
- `POST /v1/auth/refresh` - Refresh access token
- `POST /v1/auth/logout` - Invalidate session

### User Account
- `GET /v1/users/me` - Get current user details
- `DELETE /v1/users/me` - Delete account (soft delete)

### Profiles
- `GET /v1/profiles/:id` - Get user profile
- `PUT /v1/profiles/me` - Create/update own profile
- `GET /v1/profiles/me/photos` - List profile photos
- `POST /v1/profiles/me/photos` - Upload photo
- `DELETE /v1/profiles/me/photos/:photoID` - Delete photo

### Matching
- `GET /v1/matching/candidates` - Get potential matches (PostGIS-powered)
- `POST /v1/matching/like` - Like a user
- `POST /v1/matching/pass` - Pass on a user
- `GET /v1/matches` - List mutual matches

### Settings
- `GET /v1/settings` - Get user settings
- `PATCH /v1/settings` - Update settings

### Interactions & Trust
- `POST /v1/interactions` - Submit rating after meeting
- `POST /v1/interactions/:id/confirm` - Confirm an interaction
- `GET /v1/users/:id/reputation` - Get user's trust score

### Social Feed
- `GET /v1/posts` - List feed posts
- `POST /v1/posts` - Create post
- `GET /v1/posts/:id` - Get single post
- `DELETE /v1/posts/:id` - Delete post
- `POST /v1/posts/:id/like` - Like post
- `DELETE /v1/posts/:id/like` - Unlike post
- `POST /v1/posts/:id/comments` - Add comment
- `GET /v1/posts/:id/comments` - List comments

### Chat
- `GET /v1/matches/:id/messages` - Get message history
- `GET /v1/ws` - WebSocket endpoint for real-time messaging

### KYC
- `POST /v1/kyc/submit` - Submit verification document
- `GET /v1/kyc/status` - Get verification status

### Reports
- `POST /v1/reports` - Report a user

### Health
- `GET /v1/health` - Service health check

## Testing

```bash
# Run all tests
make test

# Run with verbose output
go test ./internal/... -v -race -count=1

# Run specific package tests
go test ./internal/service/... -v
```

## Environment Variables

See `.env.example` for all required variables. Key settings:

| Variable | Description |
|----------|-------------|
| `SERVER_PORT` | API server port (default: 8080) |
| `SERVER_ENV` | Environment: development/production |
| `CORS_ORIGINS` | Comma-separated allowed origins |
| `POSTGRES_*` | PostgreSQL connection settings |
| `NEO4J_*` | Neo4j connection settings |
| `REDIS_*` | Redis connection settings |
| `MINIO_*` | MinIO connection settings |
| `JWT_SECRET` | JWT signing key (32+ chars) |
| `ENCRYPTION_KEY` | AES-256 key for PII (64 hex chars) |

## Architecture

See `docs/ARCHITECTURE.md` for detailed architecture documentation covering:
- Clean architecture layers
- Trust score computation (weighted PageRank)
- Sybil detection (Louvain community detection)
- Data sovereignty compliance

## License

Proprietary - All rights reserved
