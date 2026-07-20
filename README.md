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

make docker-up     # start postgres, redis, milvus, prometheus, grafana (+ api once built)
make migrate-up    # apply database migrations
make run           # run the API locally against the containers above
```

Health checks:

```bash
curl localhost:8080/health/live
curl localhost:8080/health/ready
curl localhost:8080/health/startup
curl localhost:8080/metrics
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
| Metrics | Prometheus (client_golang) |
| Dashboards | Grafana (provisioned) |
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

Milestone 2 — Instagram Integration complete. Connect an Instagram
Business/Creator account via the Instagram Login OAuth2 flow, with
long-lived access tokens encrypted at rest (AES-256-GCM), stateless
HMAC-signed OAuth state, token refresh, and connection status.

```
GET    /api/v1/instagram/connect              start OAuth, returns authorize URL + state
GET    /api/v1/instagram/callback             OAuth redirect: exchange code, store account
GET    /api/v1/instagram/accounts             list connected accounts and status
POST   /api/v1/instagram/accounts/:id/refresh refresh a connection's long-lived token
DELETE /api/v1/instagram/accounts/:id         disconnect an account
```

Milestone 3 — Content Management complete. Full content lifecycle CRUD
with drafts, archiving, tags, and cursor-paginated listing filtered by
status, language, type, and tag.

```
POST   /api/v1/contents               create content (starts as a draft)
GET    /api/v1/contents               list own content (filters + ?limit &?cursor)
GET    /api/v1/contents/:id           fetch one content item
PATCH  /api/v1/contents/:id           update fields, tags, or draft/archived status
DELETE /api/v1/contents/:id           archive content
POST   /api/v1/contents/:id/duplicate duplicate as a new draft
```

Milestone 4 — AI Engine complete. Caption and hashtag generation through a
provider-agnostic `LLMProvider` port with an OpenAI-compatible adapter
(works with Groq, OpenRouter, or any `/v1/chat/completions` endpoint),
wired into content as a generate-caption endpoint.

```
POST /api/v1/contents/:id/caption   generate a caption + hashtags and save them
```

Configure the provider via `AI_BASE_URL` / `AI_MODEL` / `AI_API_KEY` (Groq by
default; see configs/.env.example for the OpenRouter variant).

Milestone 5 — Publishing complete. Register media on a content item, then
publish it to a connected Instagram account through the Graph Content
Publishing API (image, carousel, or reel — inferred or chosen), with a
persisted publication history.

```
POST   /api/v1/contents/:id/media                register a media URL on a content item
GET    /api/v1/contents/:id/media                list a content item's media
DELETE /api/v1/contents/:id/media/:mediaId       remove a media entry
POST   /api/v1/contents/:id/publish              publish to a connected Instagram account
GET    /api/v1/contents/:id/publications         publishing history for a content item
```

Publishing is synchronous (background workers arrive in a later milestone) and
needs live Meta credentials; media is referenced by public URL — Instagram's
Content Publishing API fetches it directly.

Milestone 6 — Scheduler complete. Schedule a content item to publish at a
future time or on a recurring cron schedule (timezone-aware, with retry
backoff). An in-process runner polls for due schedules and drives the publish
flow automatically.

```
POST   /api/v1/contents/:id/schedule              schedule a publish (scheduled_for or cron_expression + timezone)
GET    /api/v1/contents/:id/schedules             list a content item's schedules
DELETE /api/v1/contents/:id/schedules/:scheduleId cancel a pending schedule
```

The runner is in-process (poll interval `SCHEDULER_POLL_INTERVAL`).

Milestone 7 — Workers complete. Publishing runs asynchronously: schedules are
handed to a Redis Streams job queue and executed by an in-process worker pool
with at-least-once delivery, retry, a dead-letter queue, and metrics.

```
GET /api/v1/worker/stats   worker throughput (processed / succeeded / retried / dead-lettered)
```

The scheduler now enqueues an `instagram.publish` job (versioned Job Contract)
instead of publishing inline; the worker pool consumes it, retries on failure
up to the job's max attempts, then dead-letters. Concurrency is set via
`WORKER_CONCURRENCY`; the queue interface is Redis-backed but replaceable.

Milestone 8 — Knowledge Base (RAG) complete. Ingest documents (chunked +
embedded), store vectors in Milvus, and retrieve relevant chunks for a query
to build AI context — the retrieval half of Retrieval-Augmented Generation.

```
POST   /api/v1/knowledge/documents      ingest a document (chunk -> embed -> index)
GET    /api/v1/knowledge/documents      list ingested documents
DELETE /api/v1/knowledge/documents/:id  remove a document and its vectors
POST   /api/v1/knowledge/query          semantic search + built context for a query
```

The embedder and vector store are behind ports: the embedder is a deterministic
local one by default (`KNOWLEDGE_EMBEDDER=local`, no API key) or an
OpenAI-compatible `/v1/embeddings` client; the vector store is in-memory by
default or Milvus (`KNOWLEDGE_VECTOR_STORE=milvus`).

Milestone 9 — Observability complete. Prometheus metrics for every subsystem,
provisioned Grafana dashboards, alert rules, request-scoped structured logging
with correlation and trace IDs that workers inherit, and Kubernetes-shaped
health probes.

```
GET /metrics         Prometheus exposition (HTTP, worker, queue, scheduler,
                     AI, publishing, knowledge, business, pool, Go runtime)
GET /health/live     liveness — process is up
GET /health/ready    readiness — postgres, redis, queue (+ milvus when enabled)
GET /health/startup  startup — configuration, migrations, dependencies
```

Every request carries a request ID, a correlation ID (`X-Correlation-ID`, echoed
and inherited), and a trace ID (`traceparent` honoured when present). Those IDs
are attached to every log line and travel inside the job contract, so a worker
logs the same `trace_id` as the HTTP request that enqueued the job. Full
OpenTelemetry export is deliberately deferred — the instrumentation is ready.

```bash
make docker-up                       # now also starts prometheus, grafana, milvus
open http://localhost:9090           # Prometheus (targets, alerts)
open http://localhost:3030           # Grafana (admin/admin), NovaFlow folder
```

Three dashboards are provisioned: **API** (throughput, latency quantiles, error
rate, pool), **Workers & Queue** (job outcomes, execution time, queue and DLQ
depth, scheduler drift), and **Business & AI** (posts published, captions
generated, schedules, documents, token usage, retrieval latency). Alert rules
cover API downtime, 5xx rate, latency budget burn, DLQ growth, queue backlog,
worker panics, and Instagram/AI failure rates. Metrics are exposed
unauthenticated on the API port for in-cluster scraping — set
`METRICS_ENABLED=false` to turn the endpoint off.

Milestone 10 — API hardening in progress. Distributed rate limiting, security
headers, CORS allowlisting, and request size and timeout limits.

Rate limiting is a Redis-backed sliding window, so quotas hold across every API
replica rather than per process. Two tiers apply, both keyed by client IP: a
strict tier on `/api/v1/auth` to blunt credential brute force, and a general
tier across `/api/v1`. Health and metrics endpoints sit outside `/api/v1` and
are never limited, so probes and scrapes cannot be throttled. Responses carry
`X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset`, always
reporting whichever tier is closest to its limit; a rejection returns `429` with
`Retry-After`.

```bash
RATE_LIMIT_REQUESTS=100        # general tier, per window
RATE_LIMIT_AUTH_REQUESTS=10    # /api/v1/auth tier
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_FAIL_OPEN=true      # serve or 503 when Redis is unreachable
```

Security headers (CSP, HSTS, `X-Frame-Options`, `X-Content-Type-Options`,
`Referrer-Policy`, `Permissions-Policy`) are set on every response. HSTS is
emitted only over HTTPS and defaults on in production. CORS stays disabled until
`SECURITY_CORS_ALLOWED_ORIGINS` names an explicit allowlist — a wildcard origin
is rejected at startup in production and whenever credentials are allowed.

Request bodies are capped by `HTTP_BODY_LIMIT`, and read, write, and idle
timeouts bound slow-client attacks (Fiber ships with none by default).
`X-Forwarded-For` is trusted for the client IP only when `HTTP_TRUSTED_PROXIES`
lists the proxies — otherwise a client could spoof its IP and sidestep its quota.
