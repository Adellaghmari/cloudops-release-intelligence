# CV claims matrix

Truthfulness rule: nothing goes on a CV because a package was mentioned, a Terraform resource exists locally, a mock test passed, or documentation describes the intent.

Allowed status values:

| Status | Meaning |
| --- | --- |
| PLANNED | Designed, not built |
| IMPLEMENTED | Code or config exists in this repository |
| TESTED | Automated tests cover the claim with meaningful assertions |
| LIVE VERIFIED | Proven in the public/production environment |

LIVE VERIFIED is the only status that supports a strong public claim such as "I deployed X on AWS."

This remains a portfolio project. It is not professional work experience.

## Platform and languages

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Go backend | LIVE VERIFIED | Public Go API serving real JSON | Public API Gateway → Go Lambda digest `ef3778d5…`. |
| Gin HTTP API | LIVE VERIFIED | Versioned REST, tests, live `/api/v1` | Public `/api/v1` health/ready/services/releases + analysis routes 200. |
| Angular frontend | LIVE VERIFIED | Public CloudFront/S3 console | `https://d34fwrlm14h6js.cloudfront.net` production SPA; lint/unit/build + public Cypress smoke. |
| TypeScript | LIVE VERIFIED | Angular app compiled and deployed | Production bundle on CloudFront (`main-*.js`). |
| RxJS | LIVE VERIFIED | HTTP/state streams in the console | Catalog/release/status pages call live API via HttpClient. |
| SCSS | LIVE VERIFIED | Structured styles, no Tailwind | Graphite/amber console styles on public SPA. |

## AWS

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| AWS Lambda | LIVE VERIFIED | Terraform-applied functions serving traffic | Digest `ef3778d5…` API+worker. |
| Amazon API Gateway | LIVE VERIFIED | Public HTTPS API | `https://8kci5uht3d.execute-api.eu-west-1.amazonaws.com`. |
| Amazon DynamoDB | LIVE VERIFIED | Production table with real items | `cloudops-prod-main` + LIVE PROJECT DATA / Northstar rows. |
| Amazon S3 | LIVE VERIFIED | Frontend origin and/or raw event objects | Private web bucket hosts SPA; Block Public Access on; OAC only. |
| Amazon CloudFront | LIVE VERIFIED | Distribution + OAC serving SPA | `E2220NQVG6GU75` serves Angular; SPA fallback OK. Default cert truthfully reports `TLSv1` (16A Terraform aligned). Custom domain/ACM optional. |
| Amazon EventBridge | LIVE VERIFIED | Custom bus receiving real events | Ingest → analysis rule → SQS (prior + dogfood). |
| Amazon SQS | LIVE VERIFIED | Worker consuming analysis queue + DLQ | Poison `149c73ef-…` completed to DLQ (ReceiveCount≥4) + alarm ALARM; labeled synthetics cleaned. |
| Amazon ECR | LIVE VERIFIED | Immutable image exists in the repository | `sha256:ef3778d5…`. |
| Amazon CloudWatch | LIVE VERIFIED | Intended log groups and DLQ alarm exist | App logs + DLQ alarm (prior). |
| AWS X-Ray | LIVE VERIFIED | Traces for API and worker | Prior successful API/worker traces. |
| AWS IAM least privilege | LIVE VERIFIED (bounded) | Applied roles with scoped policies | Deploy writes scoped (ECR/Lambda/S3/CF) + `GetInvalidation` + prefix-restricted state IAM. Plan **and** deploy/apply roles retain AWS **ReadOnlyAccess** for Terraform refresh — broad read compatibility, **not** fully least privilege. No AdministratorAccess. |
| GitHub OIDC to AWS | LIVE VERIFIED | Workflows assume role without static keys | CD + Terraform plan + Terraform apply OIDC proven (`cloudops-prod-github-deploy`, `cloudops-prod-github-plan`). |
| AWS Secrets Manager | PLANNED / MAY OMIT | Only if a real secret is required | Intentionally avoided unless necessary |

## Infrastructure, CI, DevSecOps

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Terraform | LIVE VERIFIED | `plan`/`apply` of the intended stack | Product stack live on remote S3 state (`cloudops-prod-tfstate-7d9fdd77`, `use_lockfile=true`). GitHub plan + controlled no-op apply proven; apply gate re-disabled. |
| Terraform remote state + S3 lockfile | LIVE VERIFIED | S3 backend + native lock | Migrated Part 1; lock contention proven; GitHub plan/apply acquire/release lock; idle `.tflock` absent. |
| GitHub Terraform PLAN (OIDC) | LIVE VERIFIED | Plan role + remote state from Actions | [terraform 34389687373](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34389687373) → `cloudops-prod-github-plan`, **No changes**. |
| GitHub Terraform APPLY (controlled) | LIVE VERIFIED | Gated no-op apply from Actions | [terraform 34399074308](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34399074308): deploy OIDC, same-job 0/0/0 plan, apply **0/0/0**. Gate re-set `ENABLE_TERRAFORM_APPLY=false`; negative skip [34399312899](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34399312899). |
| GitHub Actions CI | LIVE VERIFIED | Successful workflow runs on the repo | [ci 34379595516](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34379595516). |
| GitHub Actions CD | LIVE VERIFIED | OIDC + immutable ECR push from main | [cd 34379595587](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34379595587); path filters now skip Lambda on frontend/docs-only. |
| Frontend CD (S3+CF) | LIVE VERIFIED | OIDC sync + CreateInvalidation + waiter | [cd 34389391377](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34389391377): deploy OIDC + invalidation `IEJAC1M2X6OEE1YRXIV7RKLC8` **Completed** via `GetInvalidation`. |
| Trivy scanning | LIVE VERIFIED | Workflow step that can fail the build | CRITICAL gate on `ef3778d5…`. |
| Syft SBOM | LIVE VERIFIED | Generated artifact attached to release/build | CD SBOM for Lambda image. |
| Cosign keyless signing | LIVE VERIFIED | Signed ECR image verified in CI | tlog `2771504438`. |
| Linux | LIVE VERIFIED | Real Ubuntu CI + Linux OCI image | Lambda API+worker on AL2023 image. |
| Docker / OCI containers | LIVE VERIFIED | Multi-stage image, digest identity | `cloudops-prod-api@sha256:ef3778d5…`. |
| Open Policy Agent / Rego | LIVE VERIFIED | Policies execute against real release input | Public policy surface on Northstar release via live API. |

## Product capabilities

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Release Risk Engine | LIVE VERIFIED | Persisted score + contributing signals | Public release detail risk for `rel_northstar_payments_demo`. |
| Deployment Health Comparator | LIVE VERIFIED | Pre/post windows, verdict | Public health section on release detail. |
| Release correlation | LIVE VERIFIED | Explicit evidence object, not "causation" | Health correlation wording on public UI. |
| Change Impact Graph | LIVE VERIFIED | Blast radius from persisted edges | Public impact on blast release + payments dependents. |
| Release Policy Gate | LIVE VERIFIED | Versioned policy eval | Public policy section. |
| Rollback Readiness | LIVE VERIFIED | READY/PARTIAL/NOT READY/UNKNOWN | Public rollback section. |
| Release Replay | LIVE VERIFIED | Deterministic diff of two releases | Public `/replay` route. |
| Release evidence timeline | LIVE VERIFIED | Chronological events from storage | Public timeline section. |
| Event-driven analysis | LIVE VERIFIED | Ingest → EventBridge → SQS → worker | Prior + Phase 15 dogfood ingest. |
| Idempotent ingestion | LIVE VERIFIED | Duplicate event_id does not double-apply | Prior proof. |
| Public recruiter demo | LIVE VERIFIED | Unauthenticated synthetic scenarios through real logic | CloudFront SPA + live API + SYNTHETIC/LIVE labels. |
| Self-dogfooding CI evidence | LIVE VERIFIED | This repo's real SHA/run/deploy metadata visible in-product | Status page shows `9710098e…`, run `34379595587`, digest `ef3778d5…` as LIVE PROJECT DATA. |

## Testing and reliability

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Go table-driven unit tests | TESTED | Engines covered | Local `go test` suite. |
| Integration tests | TESTED | Repository/event boundaries | DynamoDB-compatible fake + processor tests. |
| Cypress E2E | LIVE VERIFIED | Recruiter demo paths | Local + public CloudFront smoke (`recruiter.cy.ts` 2 passing). |
| k6 performance test | PLANNED | Documented run + limitations | None |
| Dead letter handling | LIVE VERIFIED | SQS DLQ exists and is observable | Newest poison `149c73ef-…` + historical `not-json` reached DLQ with ReceiveCount≥4; alarm ALARM; labeled synthetics removed. |

## Explicit non-claims

Do not list these on a CV for this project:

- Kubernetes / EKS
- Python / FastAPI / React / Next.js / Azure / Foundry / RAG / LLMs
- PostgreSQL as the primary datastore
- "Scientifically proven causality"
- "Supports millions of users"
- Professional production on-call experience
- NAT Gateway, RDS, or always-on clusters
- CloudFront default-cert TLS floor reporting as TLSv1.2_2021 (AWS reports TLSv1; Terraform now aligned)
- Fully least-privilege GitHub plan/apply IAM while `ReadOnlyAccess` remains attached
- Permanent unattended Terraform apply on every push (apply is gated off by default after the controlled proof)

## Update rule

Promote a claim only when evidence matches the status definition. Demote immediately if evidence is lost or falsified.
