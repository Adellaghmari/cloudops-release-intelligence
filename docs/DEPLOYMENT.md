# Deployment

## Topology

Single region: `eu-west-1`.

| Piece | Mechanism |
| --- | --- |
| Angular SPA | S3 + CloudFront, SHA-prefixed assets, `index.html` at origin root |
| Go API | Lambda behind API Gateway HTTP API; the frontend uses the API Gateway URL directly in the public build |
| Worker | SQS-triggered Lambda, same image different `CMD` |
| Data | DynamoDB + S3 raw bucket |
| Bus | EventBridge custom bus |

## Promotion model

Portfolio default is **local → prod**. There is no mandatory staging account.

Immutable path:

1. Merge to `main`
2. Image `repo:sha-xxxxxxxx` and digest `sha256:...`
3. Terraform / AWS API updates Lambda code to that digest
4. Frontend files uploaded under `web/<sha>/`; `index.html` rewritten to those hashed bundles
5. Smoke tests hit the public URL
6. The workflow assumes the dedicated GitHub OIDC evidence-producer role and sends a SigV4-signed canonical event to the IAM-protected ingest route

Rollback of **this** product is a Lambda code restore to the previous digest plus S3/CloudFront restore of the previous `index.html` set. The Rollback Readiness engine should eventually describe that for `cloudops-api` / `cloudops-web`.

## Terraform layout (target)

```text
infra/
  versions.tf
  providers.tf
  variables.tf
  outputs.tf
  backend.tf
  iam.tf
  dynamodb.tf
  s3.tf
  ecr.tf
  lambda.tf
  apigateway.tf
  events.tf
  sqs.tf
  cloudfront.tf
  cloudwatch.tf
  budget.tf
```

No deep module maze for one environment. No one 2,000-line `main.tf`.

State: encrypted S3 backend with the native S3 lockfile (`use_lockfile = true`), created once by a documented bootstrap. Remote state and a controlled no-op GitHub apply are verified; the production apply gate remains disabled.

## Local

`go run ./cmd/api` + `ng serve` with proxy `/api` → `localhost:8080`. `APP_ENV=local` enables bounded local event ingestion and demo reset; `.env.example` makes both controls explicit. Every non-local runtime defaults both write paths to deny. The production Lambda enables event ingestion only behind the IAM-authorized API Gateway route and keeps demo reset disabled.

## Secure producer migration status

- API Gateway IAM routes, the producer role, SigV4 helper and workflow source are **IMPLEMENTED AND LOCALLY VALIDATED**.
- They are **NOT YET DEPLOYED** and must not be described as LIVE VERIFIED.
- A reviewed Terraform plan/apply must create the route and role before `AWS_EVIDENCE_PRODUCER_ROLE_ARN` and `PUBLIC_API_URL` are added as GitHub Actions variables.
- CD skips authenticated evidence posting until both variables exist; it does not fall back to anonymous writes.
- The application deployment follows producer infrastructure so the workflow never falls back to anonymous writes.

## User-owned steps (later)

Apply and public DNS/CloudFront default domain need an AWS account and, optionally, a GitHub repo. Those are Phase 12–18. Phase 0 does not require them.
