# Relay Roadmap

This roadmap describes high-level phases, milestones, and priorities for Relay. Use this as the single source of truth for what to build next and why. Keep it small, opinionated, and regularly reviewed.

## Vision
Relay is a production-oriented multi-provider AI orchestration platform: modular, cloud-native, observable, and operator-friendly. The goal is to enable safe, reliable orchestration of multiple AI providers and workflows at scale.

## Principles
- Start simple, iterate safely
- Production defaults (secure, observable, resilient)
- Keep business logic out of infrastructure primitives
- APIs and schema first: design before implementing
- Backwards-compatibility and API versioning

## Phases
### Phase 0 — Foundations (Completed)
- Project scaffolding: `cmd/`, `internal/`, configs
- Gin HTTP server, structured JSON logging, config management
- Graceful shutdown, health checks
- Dockerfile, docker-compose, Makefile, linting, hot reload
- DB layer placeholder

### Phase 1 — Core Platform (Short-term)
Goals: establish data model, auth, and provider plumbing
- Postgres integration and migrations
- Basic domain models: Users, Providers, Workspaces, Jobs
- Authentication + authorization (JWT + sessions optional)
- Providers abstraction (interface) and simple provider adapter (mock)
- Background worker processing (simple queue using Redis)
- Basic OpenAPI spec + API versioning
- Unit & integration test coverage for core subsystems

### Phase 2 — AI Orchestration (Medium-term)
Goals: add orchestrator, streaming, and provider integrations
- Provider adapters for at least two AI providers
- Orchestration engine for multi-step pipelines and retries
- Streaming responses (SSE / WebSockets) for long-running jobs
- Rate-limiting, backoff, and provider failover strategies
- Observability: metrics (Prometheus), traces (OTel), logs aggregation

### Phase 3 — Scale & Platform (Long-term)
Goals: scale and harden for production deployments
- Kubernetes manifests, Helm chart, and CI/CD pipelines
- Multi-tenant isolation and RBAC model
- Horizontal scaling of workers, autoscaling policies
- Secrets management integration (Azure Key Vault / Vault)
- Advanced provider features: batching, caching, cost monitoring

### Phase 4 — Ecosystem & UX (Future)
- Frontend SDKs and CLI tools
- Provider marketplace & plugin system
- Policy engine for governance (safety & compliance)
- Enterprise features: audit logs, SSO, quotas, billing

## Prioritization & Criteria
- Prioritize features that unblock production use (DB, auth, provider basics)
- Measure risk and value: prefer high-value, low-risk work early
- Keep development feedback loops short (fast CI, local compose)

## How to propose roadmap changes
- Open an RFC in `docs/rfcs/` with motivation, alternatives, and migration plan
- Get 1 peer + 1 maintainer approval for non-trivial changes
- Timebox experiments; roll back if they don't prove value

## Milestones (next 90 days suggested)
1. DB + migrations + CI integration
2. Authentication + user model + protected endpoints
3. Provider interface + mock adapter + basic orchestration loop
4. Worker queue using Redis + simple job processing

---

_Last updated: 2026-06-06_
