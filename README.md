# CloudOps Release Intelligence

**Change. Risk. Impact. Recovery.**

A release decision system — not a deployment dashboard.

Most delivery tools can tell an engineer that version `1.4.2` shipped. CloudOps Release Intelligence correlates code changes, CI evidence, deployment metadata, service dependencies, health signals, policy gates, and rollback prerequisites into one inspectable release decision surface.

It answers:

- What changed?
- How risky was the release before deployment?
- Did system health deteriorate after deployment?
- Which services may be affected?
- What evidence connects a regression to this release?
- Which release policies passed, warned, or blocked?
- Is rollback operationally possible?

Where the system identifies temporal or statistical relationships it uses operational language such as **likely release correlation**, **post-deployment regression**, **risk signal**, and **potential impact**. It does not claim scientifically proven causality.

## Product thesis

Delivery telemetry is usually fragmented across Git, CI, deploy tooling, metrics, and incident systems. The missing product is a deterministic engine that turns those fragments into a decision: continue, hold, approve manually, or prepare rollback.

The innovation layer is six capabilities, each backed by persisted data and testable logic:

1. Release Risk Engine
2. Deployment Health Comparator
3. Change Impact Graph
4. Release Policy Gate
5. Rollback Readiness
6. Release Replay

## Stack

This project is deliberately a different engineering domain from an AI/RAG/Azure portfolio piece.

| Layer | Choice |
| --- | --- |
| Frontend | Angular, TypeScript, RxJS, SCSS |
| Backend | Go, Gin |
| Cloud | AWS (Lambda, API Gateway, DynamoDB, S3, EventBridge, SQS, CloudFront, CloudWatch, X-Ray, ECR, IAM) |
| IaC | Terraform |
| CI/CD | GitHub Actions with AWS OIDC |
| Policy | Open Policy Agent / Rego, embedded in Go |
| Security scanning | Trivy; Syft SBOM; Cosign keyless signing on images |
| Tests | Go testing, Angular tests, Cypress, k6 |

No LLMs. No vector databases. "Intelligence" here means deterministic operational analysis.

## Current status

**Phase 7 — Change Impact Graph (local).** Phases 0–7 are complete locally. There is no production AWS, no EventBridge/SQS execution, and no Terraform apply.

See:

- [PROJECT_STATUS.md](PROJECT_STATUS.md)
- [CV_CLAIMS_MATRIX.md](CV_CLAIMS_MATRIX.md)
- [docs/PRODUCT.md](docs/PRODUCT.md)
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)

## Architecture in one paragraph

A public Angular console is served from S3 behind CloudFront. The Go API runs as a Lambda behind API Gateway. Ingested events are written to DynamoDB (canonical state) and S3 (immutable raw payloads), then published to EventBridge. An SQS-backed analysis worker computes risk, health comparison, blast radius, policy, and rollback readiness. GitHub Actions authenticates to AWS with short-lived OIDC credentials. There is no NAT Gateway, no EKS, no RDS, and no always-on compute.

## Documentation

| Document | Purpose |
| --- | --- |
| [docs/PRODUCT.md](docs/PRODUCT.md) | Product identity, demo domain, recruiter narrative |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | System design, boundaries, API shape |
| [docs/RELEASE_RISK_ENGINE.md](docs/RELEASE_RISK_ENGINE.md) | Risk model v1 |
| [docs/HEALTH_COMPARISON.md](docs/HEALTH_COMPARISON.md) | Pre/post health comparator v1 |
| [docs/CHANGE_IMPACT_GRAPH.md](docs/CHANGE_IMPACT_GRAPH.md) | Dependency graph and blast radius |
| [docs/POLICY_AS_CODE.md](docs/POLICY_AS_CODE.md) | OPA/Rego release gates |
| [docs/ROLLBACK_READINESS.md](docs/ROLLBACK_READINESS.md) | Rollback assessment model |
| [docs/EVENT_ARCHITECTURE.md](docs/EVENT_ARCHITECTURE.md) | Typed events, idempotency, DLQ |
| [docs/DYNAMODB_DESIGN.md](docs/DYNAMODB_DESIGN.md) | Access patterns and table design |
| [docs/CI_CD.md](docs/CI_CD.md) | GitHub Actions and OIDC |
| [docs/SECURITY.md](docs/SECURITY.md) | IAM, validation, secrets, scanning |
| [docs/OBSERVABILITY.md](docs/OBSERVABILITY.md) | CloudWatch and X-Ray |
| [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) | Environments and promotion |
| [docs/DEMO.md](docs/DEMO.md) | Public recruiter scenarios |
| [docs/FAILURE_MODES.md](docs/FAILURE_MODES.md) | Chaos and graceful failure |
| [docs/COST.md](docs/COST.md) | Cost model and excluded services |
| [docs/INTERVIEW_NOTES.md](docs/INTERVIEW_NOTES.md) | Engineering stories |

## Local development

Requires Go 1.27+. No AWS credentials.

```text
go test ./...
go run ./cmd/api
```

Frontend (Angular 21, Node 22.14 — Angular 22 requires a newer Node than this machine has):

```text
cd frontend
npm start
```

The Angular dev server proxies `/api` to `http://localhost:8080`. Do not use frontend fixture JSON as API data.

Then:

```text
GET http://localhost:8080/api/v1/health
GET http://localhost:8080/api/v1/services
GET http://localhost:8080/api/v1/services?source=synthetic
GET http://localhost:8080/api/v1/releases/rel_northstar_payments_demo
POST http://localhost:8080/api/v1/events
```

Copy `.env.example` to `.env` for local overrides. Do not put credentials in git.

Local seed includes Northstar Commerce rows with `source=synthetic` and this product's own service identities with `source=live`. Phase 1 does not invent live releases.

## Cost posture

Target idle cost is a few dollars per month. Serverless on-demand services only. AWS Budgets will alarm before anything expensive can run away. See [docs/COST.md](docs/COST.md).

## License

Portfolio project. License will be set when the public GitHub repository is created.
