# Deployment

## Topology

Single region: `eu-west-1`.

| Piece | Mechanism |
| --- | --- |
| Angular SPA | S3 + CloudFront, SHA-prefixed assets, `index.html` at origin root |
| Go API | Lambda URL via API Gateway HTTP API, CloudFront `/api/*` behavior |
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
6. Evidence event ingested for dogfood

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

State: S3 backend + DynamoDB lock, created once by a bootstrap that is documented. Bootstrap is the only chicken-egg. Until AWS login exists, Terraform is written but not applied.

## Local

`go run ./cmd/api` + `ng serve` with proxy `/api` → `localhost:8080`. Worker can run as a goroutine in local mode or a second process polling a fake queue. Local mode must be explicit (`APP_ENV=local`) so production never silently no-ops AWS.

## User-owned steps (later)

Apply and public DNS/CloudFront default domain need an AWS account and, optionally, a GitHub repo. Those are Phase 12–18. Phase 0 does not require them.
