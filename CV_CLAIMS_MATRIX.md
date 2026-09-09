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
| Amazon CloudFront | LIVE VERIFIED | Distribution + OAC serving SPA | `E2220NQVG6GU75` / `d34fwrlm14h6js.cloudfront.net` serves Angular; SPA 403/404→index.html. Default cert still reports MinProto `TLSv1` despite Terraform desire `TLSv1.2_2021`. |
| Amazon EventBridge | LIVE VERIFIED | Custom bus receiving real events | Ingest → analysis rule → SQS (prior + dogfood). |
| Amazon SQS | LIVE VERIFIED | Worker consuming analysis queue + DLQ | Mapping Enabled; historical DLQ poison proven; newest poison full wait not finished in-session. |
| Amazon ECR | LIVE VERIFIED | Immutable image exists in the repository | `sha256:ef3778d5…`. |
| Amazon CloudWatch | LIVE VERIFIED | Intended log groups and DLQ alarm exist | App logs + DLQ alarm (prior). |
| AWS X-Ray | LIVE VERIFIED | Traces for API and worker | Prior successful API/worker traces. |
| AWS IAM least privilege | IMPLEMENTED | Applied roles with scoped policies | Workload path sufficient; deploy role still has broad `Resource:"*"`. |
| GitHub OIDC to AWS | LIVE VERIFIED | Workflows assume role without static keys | CD OIDC proven for Lambda and frontend deploy (`cloudops-prod-github-deploy`). |
| AWS Secrets Manager | PLANNED / MAY OMIT | Only if a real secret is required | Intentionally avoided unless necessary |

## Infrastructure, CI, DevSecOps

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Terraform | LIVE VERIFIED | `plan`/`apply` of the intended stack | Stack applied; TLS apply attempted; residual CF default-cert TLS reporting drift. Local state; GitHub apply DISABLED. |
| GitHub Actions CI | LIVE VERIFIED | Successful workflow runs on the repo | [ci 34379595516](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34379595516). |
| GitHub Actions CD | LIVE VERIFIED | OIDC + immutable ECR push from main | [cd 34379595587](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34379595587); path filters now skip Lambda on frontend/docs-only. |
| Frontend CD (S3+CF) | LIVE VERIFIED | OIDC sync + CreateInvalidation from Actions | [cd 34384226392](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34384226392): build/OIDC/S3/CreateInvalidation; waiter blocked on missing `GetInvalidation` (deferred TF). First upload was manual admin. |
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
| Dead letter handling | LIVE VERIFIED | SQS DLQ exists and is observable | Historical DLQ ReceiveCount=4 + alarm; newest poison full cycle not finished in-session. |

## Explicit non-claims

Do not list these on a CV for this project:

- Kubernetes / EKS
- Python / FastAPI / React / Next.js / Azure / Foundry / RAG / LLMs
- PostgreSQL as the primary datastore
- "Scientifically proven causality"
- "Supports millions of users"
- Professional production on-call experience
- NAT Gateway, RDS, or always-on clusters
- GitHub Terraform apply / remote state (not live)
- CloudFront default-cert TLS floor reporting as TLSv1.2_2021 (AWS still reports TLSv1)

## Update rule

Promote a claim only when evidence matches the status definition. Demote immediately if evidence is lost or falsified.
