# Product

CloudOps Release Intelligence is a **release decision system**.

It is not a CI dashboard, not a metrics explorer, and not a ticket tracker with DevOps labels.

## Identity

| Field | Value |
| --- | --- |
| Product name | CloudOps Release Intelligence |
| Tagline | Change. Risk. Impact. Recovery. |
| Core capability | Release Causality Engine (operational correlation, not scientific proof) |
| Public audience | Recruiters and hiring engineers; unauthenticated demo |
| Data | Synthetic Northstar Commerce scenarios + labeled real metadata from this project's own pipeline |

## Problem

After a deployment, an engineer usually has to assemble the story by hand:

1. Git: what files changed
2. CI: did tests and scans pass
3. Deploy: when and where it shipped
4. Metrics: did error rate or latency move
5. Dependencies: who else might break
6. Policy: was this allowed
7. Recovery: can we roll back

Those systems do not share a release-shaped evidence object. The product's job is to build that object and keep it inspectable.

## What the product is not

- Not magical causal proof. Temporal + graph + statistical evidence can support **likely release correlation**. That is not causation.
- Not an LLM product. Scoring, comparison, graph traversal, and policy evaluation are deterministic.
- Not auto-rollback. Rollback readiness is an assessment, not an actuator.
- Not a generic admin CRUD template.

## Demo organization: Northstar Commerce

A synthetic digital commerce platform with enough dependency structure to make blast radius meaningful.

| Service ID | Name | Criticality | Role |
| --- | --- | --- | --- |
| `web-storefront` | Web Storefront | HIGH | Public storefront BFF |
| `checkout-api` | Checkout API | CRITICAL | Checkout orchestration |
| `payments-service` | Payments Service | CRITICAL | Capture and authorize |
| `inventory-service` | Inventory Service | HIGH | Availability and reservations |
| `customer-api` | Customer API | HIGH | Accounts and profiles |
| `fulfillment-api` | Fulfillment API | HIGH | Post-order fulfillment |
| `notification-worker` | Notification Worker | MODERATE | Email/SMS/push worker |

This project's own workloads are modeled as first-class services and labeled **LIVE**:

| Service ID | Name | Criticality | Source |
| --- | --- | --- | --- |
| `cloudops-api` | CloudOps API | HIGH | Real Go API deployments |
| `cloudops-web` | CloudOps Web | MODERATE | Real Angular deployments |

UI and API responses must distinguish `source: synthetic` from `source: live`.

## Dependency shape

Directed edge `A → B` means **A depends on B** (A calls or consumes B).

```text
web-storefront → checkout-api
web-storefront → customer-api
web-storefront → inventory-service
checkout-api   → payments-service
checkout-api   → customer-api
checkout-api   → inventory-service
checkout-api   → fulfillment-api
payments-service → notification-worker
fulfillment-api  → inventory-service
fulfillment-api  → notification-worker
```

A change to `payments-service` has a blast radius that includes `checkout-api` (direct dependent) and `web-storefront` (transitive dependent). That result must come from graph traversal, not a hardcoded UI string.

## Recruiter surfaces

The console should make the product obvious in about ten seconds.

| Route intent | Purpose |
| --- | --- |
| Operations overview | In-flight releases, worst health deltas, blocked policies |
| Releases | List with risk, health verdict, policy, rollback |
| Release detail | Single evidence object |
| Risk analysis | Score + contributing signals |
| Health comparison | Baseline vs post-deploy windows |
| Change impact graph | Blast radius visualization |
| Policy gates | Which rule fired, version, input |
| Rollback readiness | Prerequisites checklist |
| Release replay | Deterministic A vs B |
| Event timeline | Chronological evidence |
| Services | Catalog + dependencies |
| Incidents | Linked to releases/services |
| Architecture | Honest system diagram of this product |
| System status | API, worker, queue, demo seed health |

Top-level copy:

> Most deployment tools tell you what shipped. This system connects the change to risk, service health, potential impact, and recovery readiness.

## Demo scenarios

All scenarios must pass through real application logic. Frontend fixtures must not contain final verdicts.

See [DEMO.md](DEMO.md) for the six public scenarios:

- `SAFE_RELEASE`
- `RISKY_DATABASE_RELEASE`
- `POST_DEPLOY_REGRESSION`
- `DEPENDENCY_BLAST_RADIUS`
- `SECURITY_BLOCK`
- `ROLLBACK_NOT_READY`

Timestamps are generated relative to "now" at seed time and stored as UTC.

## Human decisions

The model includes an explicit `ReleaseDecision` recorded by a human (or the demo seeder acting as a release manager):

- `PROCEED`
- `HOLD`
- `MANUAL_APPROVE`
- `ROLLBACK_PREPARE`

The system never pretends the score made the decision. It records the decision next to the evidence.
