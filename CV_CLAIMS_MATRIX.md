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
| Go backend | TESTED | Public Go API serving real JSON | Local `cmd/api` + `go test ./...`. Not publicly deployed. |
| Gin HTTP API | TESTED | Versioned REST, tests, live `/api/v1` | Local `/api/v1` health/ready/services/releases + httptest. Not live. |
| Angular frontend | TESTED | Public CloudFront/S3 console | Angular 21 local console + production build + unit tests. Not publicly deployed. Angular 22 needs Node ≥22.22.3; this machine has 22.14.0. |
| TypeScript | TESTED | Angular app compiled and deployed | `ng build` succeeds locally. |
| RxJS | TESTED | HTTP/state streams in the console | Catalog/release pages use HttpClient + RxJS. |
| SCSS | TESTED | Structured styles, no Tailwind | Global + component SCSS, no Tailwind. |

## AWS

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| AWS Lambda | PARTIAL LIVE | Terraform-applied functions serving traffic | Digest `b2569a04…` deployed: InvalidEntrypoint gone; worker empty-SQS LIVE. API exits on seed because prior image injected DynamoDB `local`/`local` credentials. Credential-chain fix is in source; awaiting next image apply. |
| Amazon API Gateway | IMPLEMENTED | Public HTTPS API | `https://8kci5uht3d.execute-api.eu-west-1.amazonaws.com` returns 500 until credential-chain image is applied. |
| Amazon DynamoDB | TESTED | Production table with real items | Table ACTIVE; ItemCount 0. Client credential-chain fix prepared; not LIVE VERIFIED for data. |
| Amazon S3 | LIVE VERIFIED | Frontend origin and/or raw event objects | Buckets `cloudops-prod-web-7be25877` and `cloudops-prod-raw-7be25877`. SPA/raw objects not uploaded. |
| Amazon CloudFront | LIVE VERIFIED | Distribution + OAC in the account | `E2220NQVG6GU75` / OAC. Angular not uploaded. Post-apply TLS floor drift may remain (`TLSv1` vs planned `TLSv1.2_2021`). |
| Amazon EventBridge | IMPLEMENTED | Custom bus receiving real events | Bus/rule/target exist. Domain processing not proven (blocked by Dynamo credentials on live image). |
| Amazon SQS | PARTIAL LIVE | Worker consuming analysis queue + DLQ | Mapping Enabled. Empty-batch worker invoke proven; real queue message domain processing not proven. |
| Amazon ECR | LIVE VERIFIED | Immutable image exists in the repository | Live `sha-35f35f20…` / `sha256:b2569a04…`. Historical `08ab5782…` retained. Next digest pending credential-chain push. |
| Amazon CloudWatch | LIVE VERIFIED | Intended log groups and DLQ alarm exist | Application logs show bootstrap success + seed failure; no secrets observed in sampled API logs. |
| AWS X-Ray | PARTIAL LIVE | Traces for API and worker | Successful worker empty-invoke trace `1-6aa18d4e-5cd76c662800c7b051f6373b`. API Gateway HTTP success traces not LIVE VERIFIED. |
| AWS IAM least privilege | IMPLEMENTED | Applied roles with scoped policies | API DynamoDB IAM includes PutItem; failure is client credentials, not missing IAM. Deploy role still has broad `Resource:"*"`. |
| GitHub OIDC to AWS | LIVE VERIFIED | Workflows assume role without static keys | CD OIDC assume-role proven. |
| AWS Secrets Manager | PLANNED / MAY OMIT | Only if a real secret is required | Intentionally avoided unless necessary |

## Infrastructure, CI, DevSecOps

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Terraform | LIVE VERIFIED | `plan`/`apply` of the intended stack | Runtime-fix apply `tfplan-runtime-fix-1`: `0 added, 3 changed, 0 destroyed`. Post-apply drift plan shows `0/2/0` (CF TLS / S3 policy); not applied. Local state; GitHub apply disabled. |
| GitHub Actions CI | LIVE VERIFIED | Successful workflow runs on the repo | [ci 34377292366](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34377292366) incl. Lambda bootstrap + RIE proof. |
| GitHub Actions CD | LIVE VERIFIED | OIDC + immutable ECR push from main | [cd 34377292422](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34377292422) pushed `sha-35f35f20…` / digest `b2569a04…`. |
| Trivy scanning | LIVE VERIFIED | Workflow step that can fail the build | CRITICAL gate passed on `b2569a04…`. Residual 2 HIGH openssl on AL2023 base (not suppressed). |
| Syft SBOM | LIVE VERIFIED | Generated artifact attached to release/build | `cloudops-lambda-sbom.spdx.json` from CD 34377292422. |
| Cosign keyless signing | LIVE VERIFIED | Signed ECR image verified in CI | Sign+verify succeeded for `b2569a04…` (tlog `2771300046`). |
| Linux | PARTIAL LIVE | Real Ubuntu CI + Linux OCI image | CI Ubuntu + RIE proven. Worker empty-invoke on AWS proven. API serving traffic blocked by Dynamo client credentials. |
| Docker / OCI containers | LIVE VERIFIED | Multi-stage image, non-root, digest identity | Deployed `cloudops-prod-api@sha256:b2569a0452d2be9c1be0e509db7dea7bc0a12d202e714334702cb2de1985c40f`. |
| Open Policy Agent / Rego | TESTED | Policies execute against real release input | `policies/release_gate.rego` evaluated in-process via opa/v1/rego. Local only. |

## Product capabilities

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Release Risk Engine | TESTED | Persisted score + contributing signals from real logic | `internal/risk` + GET /releases/:id/risk. Local only. |
| Deployment Health Comparator | TESTED | Pre/post windows, raw metrics, verdict | `internal/health` + GET /releases/:id/health. Local only. Not CloudWatch verified. |
| Release correlation | TESTED | Explicit evidence object, not "causation" | LIKELY/POSSIBLE/NO_CLEAR/INSUFFICIENT_DATA with reasons. Never PROVEN_CAUSE. Local only. |
| Change Impact Graph | TESTED | BFS/DFS blast radius from persisted edges | `internal/graph` + GET /releases/:id/impact + SVG. Language is potential impact. Local only. |
| Release Policy Gate | TESTED | Versioned policy eval with PASS/WARN/BLOCK/MANUAL | Real Rego tests + GET /releases/:id/policy. Fail closed. Local only. |
| Rollback Readiness | TESTED | READY/PARTIAL/NOT READY/UNKNOWN + missing prereqs | `internal/rollback` + GET /releases/:id/rollback. No auto rollback. Local only. |
| Release Replay | TESTED | Deterministic diff of two persisted releases | GET /replay + Angular Replay page. Local only. |
| Release evidence timeline | TESTED | Chronological events from storage | GET /releases/:id/timeline from persisted events. Local only. |
| Event-driven analysis | TESTED | Ingest → EventBridge → SQS → worker, not in-request theatre | Local processor still TESTED. AWS bus/rule/queue/DLQ exist; worker and event source mapping do not. Not LIVE VERIFIED. |
| Idempotent ingestion | TESTED | Duplicate event_id does not double-apply | Processor + memory/DynamoDB conditional CreateEvent + HTTP 200 duplicate. Local only. |
| Public recruiter demo | IMPLEMENTED | Unauthenticated synthetic scenarios through real logic | Local console + `POST /demo/reset`. Not publicly deployed. |
| Self-dogfooding CI evidence | IMPLEMENTED | This repo's real SHA/run/deploy metadata visible in-product | Real metadata now exists: git SHA `c22171c3b6b8bd3946a2972f16283eac629f5ce3`, CI run `34370805926`, CD run `34370805611`, image digest `sha256:08ab57825cc1f43a1527ad474f20a146e8f55dacb2c49323f42601129efa1a53`. Not ingested because the public API does not exist. Keep LIVE PROJECT DATA separate from SYNTHETIC DEMO. |

## Testing and reliability

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Go table-driven unit tests | TESTED | Engines covered | Domain/repo/API tests passing locally. Engine packages not started. |
| Integration tests | TESTED | Repository/event boundaries | DynamoDB-compatible fake + event processor tests. Not DynamoDB Local / LocalStack. |
| Cypress E2E | TESTED | Recruiter demo paths | `frontend/cypress/e2e/recruiter.cy.ts` against local API+Angular. Not public. |
| k6 performance test | PLANNED | Documented run + limitations | None |
| Dead letter handling | TESTED | SQS DLQ exists and is observable | In-process Processor DLQ TESTED. AWS queue `cloudops-prod-analysis-dlq` and alarm exist; poison path not exercised by a worker. |

## Explicit non-claims

Do not list these on a CV for this project:

- Kubernetes / EKS
- Python / FastAPI / React / Next.js / Azure / Foundry / RAG / LLMs
- PostgreSQL as the primary datastore
- "Scientifically proven causality"
- "Supports millions of users"
- Professional production on-call experience
- NAT Gateway, RDS, or always-on clusters

## Update rule

Update this matrix at the end of every phase. Promote a row only when the evidence column can name a concrete artifact: test name, workflow run URL, Terraform apply, public URL, or CloudWatch log group.
