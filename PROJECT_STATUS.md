# Project status

**Project:** CloudOps Release Intelligence
**Phase:** 1 — Core domain models + Go API
**Status:** PHASE 1 COMPLETE
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
- Local catalog seed (synthetic Northstar + live service identities, no invented live releases)
- Versioned read APIs for health, readiness, services, releases

Not implemented: DynamoDB, EventBridge, SQS, risk/health/graph/policy/rollback engines, Angular, Terraform, GitHub Actions.

## Test status

| Suite | Status |
| --- | --- |
| Go unit tests | PASSING (local) |
| Go integration tests | NOT STARTED (no DynamoDB yet) |
| Angular unit tests | NOT STARTED |
| Cypress E2E | NOT STARTED |
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
- No AWS resources. Readiness reports `in_memory_store` only.
- No ingestion API, webhooks, or async worker.
- No risk/health/graph/policy/rollback computation.
- No Angular console.
- Module path `github.com/adell/cloudops-release-intelligence` is a placeholder until the public remote exists.
- Go toolchain installed for this phase is **1.27.0** (current major). Winget did not offer 1.27.1.
- Default local git branch is still `master` until renamed; CI/CD will target `main`.

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

**Phase 2 — DynamoDB + event persistence** (local/dev first, still no production AWS apply): implement the same `repository.Store` against DynamoDB (or DynamoDB local), persist events with `event_id` conditional writes, keep domain types free of PK/SK.

## Phase tracker

| Phase | Name | Status |
| --- | --- | --- |
| 0 | Architecture and repository foundation | COMPLETE |
| 1 | Core domain models + Go API | COMPLETE |
| 2 | DynamoDB + event persistence | PLANNED |
| 3 | Angular operations console | PLANNED |
| 4 | Release Risk Engine | PLANNED |
| 5 | EventBridge + SQS asynchronous processing | PLANNED |
| 6 | Deployment Health Comparator | PLANNED |
| 7 | Change Impact Graph | PLANNED |
| 8 | OPA / Release Policy Gate | PLANNED |
| 9 | Rollback Readiness | PLANNED |
| 10 | Release Timeline + Release Replay | PLANNED |
| 11 | Real GitHub Actions / self dogfooding | PLANNED |
| 12 | AWS infrastructure with Terraform | PLANNED |
| 13 | CI/CD + OIDC + DevSecOps | PLANNED |
| 14 | CloudWatch / X-Ray observability | PLANNED |
| 15 | Public synthetic recruiter demo | PLANNED |
| 16 | Security / reliability / performance hardening | PLANNED |
| 17 | Full live verification | PLANNED |
| 18 | GitHub repository finalization | PLANNED |
