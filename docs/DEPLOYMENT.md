# Deployment

## Topology

Single region: `eu-west-1`.

| Piece | Mechanism |
| --- | --- |
| Angular SPA | S3 + CloudFront, SHA-prefixed assets, `index.html` at origin root; optional `frontend_custom_domain` + `frontend_acm_certificate_arn` (us-east-1) when you own DNS |
| Go API | Lambda behind API Gateway HTTP API; the frontend uses the API Gateway URL directly in the public build |
| Worker | SQS-triggered Lambda, same image different `CMD` |
| Data | DynamoDB + S3 raw bucket |
| Bus | EventBridge custom bus |

## Promotion model

Portfolio default is **local → prod**. There is no mandatory staging account.

Immutable path:

1. Merge to `main`
2. Manually dispatch the protected release for that exact `main` SHA
3. Build, scan, and publish image `repo:sha-<full-sha>` once, then resolve its immutable digest
4. Generate and inspect a saved Terraform plan bound to that digest
5. Obtain `prod` environment approval and apply that exact plan
6. Terraform updates both Lambda functions to the planned digest
7. Build and upload the frontend, then wait for CloudFront invalidation
8. Assume the dedicated GitHub OIDC evidence producer role and send a SigV4 signed canonical event to the IAM protected ingest route
9. Verify the live release and require a zero change Terraform drift plan

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

State: encrypted S3 backend with the native S3 lockfile (`use_lockfile = true`), created once by a documented bootstrap. The bounded production Terraform role and protected GitHub environment are live verified. Intentional releases use a reviewed saved plan; the apply kill switch remains disabled outside an approved release window.

## Local

`go run ./cmd/api` + `ng serve` with proxy `/api` → `localhost:8080`. `APP_ENV=local` enables bounded local event ingestion and demo reset; `.env.example` makes both controls explicit. Every non-local runtime defaults both write paths to deny. The production Lambda enables event ingestion only behind the IAM-authorized API Gateway route and keeps demo reset disabled.

## Secure producer migration status

- API Gateway IAM routes, the producer role, SigV4 helper and workflow source are **IMPLEMENTED AND LOCALLY VALIDATED**.
- They are **NOT YET DEPLOYED** and must not be described as LIVE VERIFIED.
- A reviewed Terraform plan and apply must create the route and role before evidence posting.
- The protected release reads the producer role ARN and API endpoint from applied Terraform outputs. No producer repository variables or static credentials are required.
- Missing outputs fail the release. There is no anonymous fallback.
- Application deployment follows producer infrastructure, and signed evidence follows application deployment.

## User-owned steps (later)

Apply and public DNS/CloudFront default domain need an AWS account and, optionally, a GitHub repo. Those are Phase 12–18. Phase 0 does not require them.
