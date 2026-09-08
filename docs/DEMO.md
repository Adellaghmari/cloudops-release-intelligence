# Public demo

Unauthenticated. Synthetic data only for Northstar Commerce. Live identities are this product's own catalog rows (`cloudops-api`, `cloudops-web`) with **no invented live releases**.

All Northstar rows are labeled **SYNTHETIC DEMO**.

## Dynamic time

Seed generates timestamps relative to seed time, stores UTC, renders local in the Angular client. Fixtures never contain the words "today" or "10 minutes ago" as stored values.

## Scenarios

Each scenario is a set of **inputs**. Verdicts come from engines. Do not treat the notes below as hardcoded UI strings.

### `SAFE_RELEASE` — `rel_northstar_payments_demo`

Small `payments-service` change, CI green, no migration, prior successful release with artifact + digest, stable health snapshots.

### `RISKY_DATABASE_RELEASE` — `rel_northstar_risky_db`

Large `checkout-api` change with `migration_present` and reversibility recorded as false.

### `POST_DEPLOY_REGRESSION` — `rel_northstar_regression`

`checkout-api` deploy with a healthy baseline and a severely worse post window, plus an open incident.

### `DEPENDENCY_BLAST_RADIUS` — `rel_northstar_blast`

`inventory-service` change. Blast radius is calculated from persisted edges (checkout, storefront, fulfillment depend on inventory).

### `SECURITY_BLOCK` — `rel_northstar_security`

`customer-api` release with a persisted CRITICAL security scan.

### `ROLLBACK_NOT_READY` — `rel_northstar_rollback`

First `fulfillment-api` release: no previous successful release and no artifact/digest metadata.

## Demo reset

`POST /api/v1/demo/reset` is still PLANNED. Locally, restarting `go run ./cmd/api` reseeds the memory store.

## Recruiter path

1. See name + tagline
2. Open Releases and a SYNTHETIC DEMO scenario
3. Read risk factors, health comparison, potential impact, policy, rollback, timeline
4. Open Replay and compare two persisted releases

## Cypress

`frontend/cypress/e2e/recruiter.cy.ts` covers the local recruiter path against the running Angular console and Go API. Not a public CloudFront verification.
