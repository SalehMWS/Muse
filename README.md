# NovaFlow

AI-powered Instagram Content Management Platform — generate captions,
schedule posts, and publish to Instagram from a single backend.

Modular Monolith · Go · Fiber · PostgreSQL · Redis · Milvus · MinIO

## Architecture

Clean Architecture, dependency-inversion, one module per business
capability. Dependencies point inward only:

```
Delivery (HTTP) → Application (use cases) → Domain (business rules)
                          ↑
                   Infrastructure (Postgres, Redis, external APIs)
```

The domain layer is pure Go — no HTTP, no SQL, no third-party SDKs. Every
external dependency (AI provider, social platform, storage) sits behind an
interface owned by its consumer, so providers are swappable without
touching business logic.

## Requirements

- Go 1.25+
- Docker & Docker Compose
- `sqlc`, `goose`, `golangci-lint` (for local development)

## Getting started

```bash
cp configs/.env.example configs/.env

make docker-up     # start postgres + redis (+ api once built)
make migrate-up    # apply database migrations
make run           # run the API locally against the containers above
```

Health checks:

```bash
curl localhost:8080/health/live
curl localhost:8080/health/ready
```

## Development

```bash
make fmt      # gofmt + goimports
make vet      # go vet
make lint     # golangci-lint
make test     # go test -race ./...
```

## Project layout

```
cmd/            entry points (api, migrate, ...)
internal/       private application code, one directory per module
  shared/       cross-cutting infrastructure: config, logger, database,
                cache, errors, response, middleware, health
configs/        environment templates
deployments/    docker-compose, Dockerfiles, database migrations
scripts/        automation scripts
tests/          integration fixtures
```

Every business capability (auth, content, instagram, ai, publishing,
scheduler, ...) owns a module under `internal/<module>/` with `domain/`,
`application/`, `infrastructure/`, `delivery/`, and `module.go`. Modules
never reach into another module's persistence layer directly — they
communicate through interfaces or events.

## Tech stack

| Layer | Choice |
|---|---|
| Language | Go |
| HTTP framework | Fiber |
| Database | PostgreSQL (pgx) |
| SQL generation | sqlc |
| Migrations | goose |
| Cache / Queue | Redis |
| Vector store | Milvus |
| Object storage | MinIO |
| Logging | zap (structured) |
| Deployment | Docker Compose → Kubernetes (future) |

## Status

Milestone 0 — Project Foundation complete (config, logger, DI bootstrap,
health checks, migrations, CI, lint).

Milestone 1 — Authentication complete. Email/password auth with Argon2id
password hashing, JWT access tokens, and rotating refresh-token sessions.

```
POST /api/v1/auth/register   create an account
POST /api/v1/auth/login      authenticate, returns access + refresh tokens
POST /api/v1/auth/refresh    rotate a refresh token for a new access token
POST /api/v1/auth/logout     invalidate a refresh token
GET  /api/v1/auth/me         current authenticated user (Bearer token)
```

Instagram integration is next.
