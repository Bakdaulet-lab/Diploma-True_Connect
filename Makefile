.PHONY: build run test lint migrate-up migrate-down docker-up docker-down clean seed-halal reset-demo

# Build all binaries
build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
	go build -o bin/migrate ./cmd/migrate

# Run the API server locally
run:
	go run ./cmd/api

# Run the worker locally
run-worker:
	go run ./cmd/worker

# Run all tests
test:
	go test ./... -v -race -count=1

# Run linter (requires golangci-lint installed)
lint:
	golangci-lint run ./...

# Database migrations (sources .env so Go picks up PG_PASSWORD etc.)
# PG_HOST is overridden to localhost because migrate runs outside Docker.
migrate-up:
	bash -c 'set -a; source .env; set +a; PG_HOST=localhost go run ./cmd/migrate up'

migrate-down:
	bash -c 'set -a; source .env; set +a; PG_HOST=localhost go run ./cmd/migrate down'

# Docker
docker-up:
	docker compose -f deployments/docker-compose.yml --env-file .env up --build -d

docker-down:
	docker compose -f deployments/docker-compose.yml down

docker-logs:
	docker compose -f deployments/docker-compose.yml logs -f

# Reset everything (volumes too)
docker-reset:
	docker compose -f deployments/docker-compose.yml down -v
	docker compose -f deployments/docker-compose.yml --env-file .env up --build -d

# Clean build artifacts
clean:
	rm -rf bin/

# Load Айгерим + Алихан demo users into the running Postgres container.
seed-halal:
	cat migrations/seed_halal_demo.sql | docker compose -f deployments/docker-compose.yml exec -T postgres psql -U postgres -d trueconnect

# Full reset: tear down volumes, rebuild, migrate, and reseed demo data.
reset-demo:
	docker compose -f deployments/docker-compose.yml down -v
	docker compose -f deployments/docker-compose.yml --env-file .env up --build -d
	@echo "Waiting for Postgres to be ready..."
	@sleep 12
	$(MAKE) migrate-up
	$(MAKE) seed-halal
