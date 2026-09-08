# Project status

**Project:** CloudOps Release Intelligence
**Phase:** 10 — Release Timeline + Release Replay
**Status:** PHASE 10 COMPLETE (local)
**Complete:** No

This file is the source of truth for what exists versus what is planned. It must not mark the project COMPLETE until the public product, AWS infrastructure, CI/CD, observability, and live verification criteria in the master build prompt are actually proven.

## Current phase

Phase 1 is complete when:

- [x] Go module and idiomatic layout exist
- [x] Typed identifiers for release, service, deployment, event, incident, and related entities
- [x] Domain entities with validation and UTC timestamps
- [x] Explicit `live` vs `synthetic` source on catalog/release/incident/health data
- [x] Repository interfaces with no AWS/HTTP types
- [x] In-memory repository
- [x] Service layer (handlers do not own business lookups)
- [x] Gin `/api/v1` health, ready, services, releases
- [x] Structured JSON errors and request/correlation IDs
- [x] Structured slog, graceful shutdown, config package
- [x] `go fmt`, `go vet`, `go test ./...` pass
- [x] Status and CV matrix updated (nothing LIVE VERIFIED)

## Implemented features

- Local Go API (`go run ./cmd/api`)
- Domain model: Service, Dependency, Release, Deployment, Commit, CIRun, HealthSnapshot, Incident, ReleaseEvent, ReleaseDecision
- In-memory persistence with duplicate-identity rejection
- DynamoDB single-table adapter (AWS SDK v2) behind `repository.Store`; `APP_STORE=memory|dynamodb`
- Conditional EventID writes; timeline query by release
- Local catalog seed (synthetic Northstar + live service identities, no invented live releases)
- Versioned read APIs for health, readiness, services, releases

- Event envelope (`schema_version` 1.0), local `Processor`, `MemoryBus`, in-process DLQ
- `POST /api/v1/events` with duplicate EventID detection (202 new / 200 duplicate)
- Release Risk Engine (`internal/risk`) + GET `/releases/:id/risk`
- Health comparator + release correlation (`internal/health`) + GET `/releases/:id/health`
- Change impact graph (`internal/graph`) + GET `/releases/:id/impact` + SVG visualization
- In-process OPA/Rego policy gate (`policies/` v1.0.0) + GET `/releases/:id/policy`
- Rollback readiness (`internal/rollback`) + GET `/releases/:id/rollback` (decision support only)
- Release timeline + replay (`internal/timeline`, `internal/replay`) + GET `/releases/:id/timeline` + GET `/replay`
- Northstar synthetic scenarios (SAFE_RELEASE, RISKY_DATABASE_RELEASE, POST_DEPLOY_REGRESSION, DEPENDENCY_BLAST_RADIUS, SECURITY_BLOCK, ROLLBACK_NOT_READY)
- Angular Release Detail sections + Replay page; Cypress recruiter path
- Angular 21 operations console (local)

Not implemented: production DynamoDB, EventBridge, SQS, LocalStack, Terraform, GitHub Actions, public demo hosting.

## Test status

| Suite | Status |
| --- | --- |
| Go unit tests | PASSING (local) |
| Go integration tests | TESTED against DynamoDB-compatible fake (not DynamoDB Local) |
| Angular unit tests | PASSING locally (Phase 3) |
| Cypress E2E | TESTED locally (recruiter path; not public) |
| k6 performance | NOT STARTED |
| Terraform validate | NOT STARTED |

## Deployment status

| Environment | Status |
| --- | --- |
| Local | IMPLEMENTED (in-memory) |
| Staging | NOT STARTED |
| Production | NOT STARTED |
| Public frontend | NOT STARTED |
| Public API | NOT STARTED |

## Known limitations

- Persistence is process-local. Restart wipes state except what `localseed` reloads.
- DynamoDB adapter is TESTED against an in-process compatible fake. Docker/Java are unavailable, so DynamoDB Local was not run.
- No AWS resources. Readiness reports `in_memory_store` only.
- EventBridge and SQS are not executed. `MemoryBus` is in-process only. Not LIVE VERIFIED.
- LocalStack was not used (Docker unavailable).
- No webhooks. GitHub signature verification is specified, not implemented.
- Public CloudFront/S3 hosting and demo reset API are not implemented.
- Module path `github.com/adell/cloudops-release-intelligence` is a placeholder until the public remote exists.
- Go toolchain is **1.27.0**. Angular is 21 (Node 22.14.0; Angular 22 needs ≥22.22.3).
- Default local git branch is `main`. No GitHub remote yet.

## Known security limitations

- Public write surface does not exist yet.
- GitHub webhook signature verification is specified, not implemented.
- OIDC trust is specified, not implemented.
- Local CORS allowlist is localhost-only by default.
- No secrets exist in the repository. Keep it that way.

## Known cost limitations

- No AWS bill yet.
- Cost estimate remains valid only if later phases keep the serverless exclusions in [docs/COST.md](docs/COST.md).

## Remaining work

**Phase 11 — Real GitHub Actions / self dogfooding** (not started; this batch stops at Phase 10).

## Phase tracker

| Phase | Name | Status |
| --- | --- | --- |
| 0 | Architecture and repository foundation | COMPLETE |
| 1 | Core domain models + Go API | COMPLETE |
| 2 | DynamoDB + event persistence | COMPLETE |
| 3 | Angular operations console | COMPLETE |
| 4 | Release Risk Engine | COMPLETE |
| 5 | Event architecture foundation (local ports) | COMPLETE (local; EventBridge/SQS not AWS verified) |
| 6 | Deployment Health Comparator + correlation | COMPLETE (local) |
| 7 | Change Impact Graph | COMPLETE (local) |
| 8 | OPA / Release Policy Gate | COMPLETE (local) |
| 9 | Rollback Readiness | COMPLETE (local) |
| 10 | Release Timeline + Release Replay | COMPLETE (local) |
| 11 | Real GitHub Actions / self dogfooding | PLANNED |
| 12 | AWS infrastructure with Terraform | PLANNED |
| 13 | CI/CD + OIDC + DevSecOps | PLANNED |
| 14 | CloudWatch / X-Ray observability | PLANNED |
| 15 | Public synthetic recruiter demo | PLANNED |
| 16 | Security / reliability / performance hardening | PLANNED |
| 17 | Full live verification | PLANNED |
| 18 | GitHub repository finalization | PLANNED |
