# Relay Architecture

This document defines the architectural decisions and boundaries for Relay. It is the "north star" used to prevent drift as the project grows.

## Goals
- Modular monolith: single deployable with clear module boundaries that can later split into services
- Testable and observable components
- Explicit dependency injection and minimal global state
- Production defaults (secure by default, run well in containers/Kubernetes)

## High-level Layers
1. cmd/ — application entry points (e.g., `cmd/api/main.go`). Minimal bootstrap only.
2. internal/ — private application packages (not intended for external import)
   - config — typed configuration and environment loading
   - logger — structured JSON logging wrapper
   - server — HTTP setup, routes, middleware
   - health — health/readiness checks
   - db — database connections and repository layer
   - providers — provider adapters (future)
   - ai / orchestrator — orchestration engine (future)
   - workers — background processing
3. migrations/ — SQL migrations tracked with a migration tool
4. configs/ — environment templates (do not commit secrets)

## Package Boundaries & Responsibilities
- Handlers (HTTP) should be thin: parse/validate input, call application services, return responses
- Application services contain business logic and coordinate repositories/infrastructure
- Repositories encapsulate database access (SQL, queries, transactions)
- Provider adapters implement a stable interface to external AI providers

Example: request flow for creating a job
1. HTTP handler receives request → validates input
2. Handler calls `JobService.CreateJob(ctx, dto)`
3. `JobService` persists a job via `JobRepository` and enqueues work in Redis
4. Worker reads job from queue and executes orchestration logic via provider adapters
5. Results are persisted; streaming clients receive updates via SSE/WebSocket

## Data Modeling & Persistence
- Primary datastore: PostgreSQL (ACID, relational model)
- Short-lived data / queues: Redis (in-memory, persistence optional)
- Use migrations (golang-migrate) for schema changes
- Prefer explicit SQL or codegen (sqlc) over heavy ORMs in early phases

## Configuration & Secrets
- 12-factor: config via environment variables
- `configs/.env.example` documents keys for local dev
- For production, use a secrets manager (Azure Key Vault, HashiCorp Vault)

## Logging, Metrics, Tracing
- Logging: structured JSON via `slog`, include correlation/request IDs
- Metrics: expose Prometheus metrics endpoint (`/metrics`) from server or a sidecar
- Tracing: integrate OpenTelemetry for distributed tracing across workers and HTTP
- Correlate logs + traces + metrics using trace IDs and request IDs

## Observability & Alerting
- Liveness and readiness endpoints for Kubernetes
- Prometheus for metrics scraping, Grafana for dashboards
- Alerts for high error rate, high queue lag, and resource saturation

## Resilience & Reliability Patterns
- Graceful shutdown for HTTP server and workers
- Retry/backoff and circuit breaker patterns for provider calls
- Rate limiting per-user and per-provider to avoid abuse
- Provider fallback strategies (e.g., try alternate provider when one fails)

## Security
- Default to `APP_ENV=production` security posture in prod manifests
- TLS in front of the service (Ingress/controller or load balancer)
- Protect sensitive endpoints with authentication and RBAC
- Use secrets manager, do not store plaintext secrets in repo
- Harden Docker images (multi-stage, non-root user)

## API Design Principles
- Version APIs (e.g., `/api/v1/...`) from day one
- Return well-structured error responses with codes and machine-readable details
- Document APIs with OpenAPI and keep spec in repo

## Deployment & CI/CD
- Build reproducible Docker images using multi-stage Dockerfile
- CI: lint, unit tests, build, and run integration tests in pipeline
- CD: push images to registry, deploy via Helm or Kubernetes manifests
- Use health checks and rolling updates for zero-downtime deploys

## Scaling Strategy
- Scale stateless HTTP horizontally (Kubernetes Deployment)
- Scale workers separately based on queue lag
- Use Redis for coordination and short-term caches
- Partition long-running jobs and orchestrator state where needed

## Tradeoffs & Decisions
- Start as a monolith to reduce operational complexity; split later if needed
- Avoid heavy ORMs early: explicit SQL provides clarity and performance
- Use `slog` for structured logging (native stdlib) — simpler integration

## Coding Conventions
- Use `context.Context` as first parameter in APIs that are request-scoped
- Constructor/DI: pass dependencies explicitly (e.g., `NewService(repo, log, cfg)`)
- Keep handlers thin and move business logic to services
- Favor small, focused packages with clear responsibilities

## Migration & Backwards Compatibility Strategy
- Migrations are applied in CI/CD with careful rollouts
- For breaking schema changes, use dual-writes or transactional migrations when possible
- Maintain API compatibility for at least one major version

## Governance
- Changes to architecture should be proposed as RFCs in `docs/rfcs/` with migration plan
- Major infra changes require maintainer approval and a rollback plan

---
_Last updated: 2026-06-06_
