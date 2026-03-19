# ============================================================
# Stage 1: Build
# ============================================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/migrate ./cmd/migrate

# ============================================================
# Stage 2: Runtime
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 1000 appuser
USER appuser

COPY --from=builder /bin/api /bin/api
COPY --from=builder /bin/worker /bin/worker
COPY --from=builder /bin/migrate /bin/migrate
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/bin/api"]
