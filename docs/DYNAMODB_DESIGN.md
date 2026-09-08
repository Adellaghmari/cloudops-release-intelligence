# DynamoDB design

DynamoDB is not SQL. This design starts from access patterns, then keys.

## Decision: one table, two GSIs

**Table:** `cloudops-main` (`PAY_PER_REQUEST`, AWS-owned encryption)

**Why single table**

- Almost every recruiter query is release-scoped or service-scoped.
- One table keeps Terraform, IAM, and cost small.
- Conditional writes for idempotency are natural.

**Why not three independent tables**

- Joins would be client-side anyway.
- Extra IAM and alarms for little isolation benefit at this size.

**Why not a classic 8-GSI kitchen sink**

- Each GSI doubles write cost. Two GSIs cover the required patterns.

Raw webhook payloads and scan reports go to **S3**, not DynamoDB, when they exceed a small excerpt.

## Keys

| Attribute | Role |
| --- | --- |
| `PK` | Partition |
| `SK` | Sort |
| `GSI1PK` / `GSI1SK` | Service and reverse-dependency queries |
| `GSI2PK` / `GSI2SK` | Type listings and global recency |
| `expires_at` | TTL, used only for idempotency records (90 days) |

## Item map

| Entity | PK | SK | GSI1PK | GSI1SK | GSI2PK | GSI2SK |
| --- | --- | --- | --- | --- | --- | --- |
| Service | `SERVICE#{id}` | `META` | `TYPE#SERVICE` | `NAME#{name}` | | |
| Dependency | `SERVICE#{from}` | `DEP#{to}` | `SERVICE#{to}` | `DEPBY#{from}` | | |
| Release | `RELEASE#{id}` | `META` | `SERVICE#{sid}` | `RELEASE#{createdAt}#{id}` | `TYPE#RELEASE` | `TIME#{createdAt}#{id}` |
| Deployment | `RELEASE#{id}` | `DEPLOY#{id}` | `SERVICE#{sid}` | `DEPLOY#{at}#{id}` | | |
| CI run | `RELEASE#{id}` | `CI#{id}` | | | | |
| Security scan | `RELEASE#{id}` | `SCAN#{id}` | | | | |
| Artifact | `RELEASE#{id}` | `ART#{id}` | | | | |
| Health snapshot | `SERVICE#{sid}` | `HEALTH#{windowEnd}#{id}` | `RELEASE#{rid}` | `HEALTH#{at}` | | |
| Health comparison | `RELEASE#{id}` | `HCOMP#LATEST` | | | | |
| Incident | `INCIDENT#{id}` | `META` | `SERVICE#{sid}` | `INCIDENT#{openedAt}` | `RELEASE#{rid}` | `INCIDENT#{openedAt}` |
| Risk | `RELEASE#{id}` | `RISK#LATEST` | | | | |
| Policy eval | `RELEASE#{id}` | `POLICY#{phase}#{id}` | | | | |
| Policy latest | `RELEASE#{id}` | `POLICY#{phase}#LATEST` | | | | |
| Rollback | `RELEASE#{id}` | `ROLLBACK#LATEST` | | | | |
| Decision | `RELEASE#{id}` | `DECISION#{id}` | | | | |
| Timeline event | `RELEASE#{id}` | `EVENT#{occurredAt}#{eventId}` | | | | |
| Idempotency | `IDEM#{eventId}` | `META` | | | | |
| Analysis idempotency | `IDEM#{eventId}#{kind}` | `META` | | | | |

Timestamps in keys are UTC ISO-8601. That keeps sort order chronological.

## Access patterns

| Need | How |
| --- | --- |
| Get one release | `GetItem PK=RELEASE#{id}, SK=META` then `Query PK=RELEASE#{id}` for children |
| Recent releases for a service | `Query GSI1 PK=SERVICE#{sid}, SK begins_with RELEASE#` scan index backward |
| Global recent releases | `Query GSI2 PK=TYPE#RELEASE` newest first |
| Events for a release | `Query PK=RELEASE#{id}, SK begins_with EVENT#` |
| Current dependencies from a service | `Query PK=SERVICE#{id}, SK begins_with DEP#` |
| Reverse dependents (blast radius BFS) | `Query GSI1 PK=SERVICE#{id}, SK begins_with DEPBY#` |
| Health around a deployment | `Query PK=SERVICE#{sid}, SK between HEALTH#{T-60m} and HEALTH#{T+32m}` and/or GSI1 on release |
| Policy evaluations | `Query PK=RELEASE#{id}, SK begins_with POLICY#` |
| Compare two releases | two Get/Query in parallel; comparison is computed, not stored as a third release |
| Idempotent ingest | `PutItem PK=IDEM#{eventId}` with `attribute_not_exists(PK)` |

## Idempotency

Ingest uses a conditional put on the idempotency item. Duplicate producers receive success without a second EventBridge emit.

Assessment pointers (`RISK#LATEST`, `HCOMP#LATEST`, `ROLLBACK#LATEST`) are overwritten with the same model version; workers may also keep the previous copy as `RISK#{assessedAt}` if useful for audit. v1 stores LATEST plus timeline events that point at the assessment.

## Size and types

- Keep items well under 400 KB.
- Store factor arrays and policy rule arrays inline (they are small).
- Store raw GitHub payloads in S3; DynamoDB holds pointer `s3_uri`.
- Numbers stored as DynamoDB `N`; booleans as `BOOL`; IDs as `S`.

## Local development

Phase 1: in-memory map implementing the same repository interface.
Phase 2: DynamoDB local via testcontainers for repository tests.
Production: real table from Terraform.

Do not use a SQL compatibility layer.
