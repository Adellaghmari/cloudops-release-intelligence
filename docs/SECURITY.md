# Security

Deliberate controls. No security theatre badges.

## Identity and access

- GitHub Actions: OIDC → IAM deploy role (see [CI_CD.md](CI_CD.md)).
- Lambda: execution roles, not embedded keys.
- DynamoDB/S3/EventBridge/SQS: resource-scoped IAM.
- Public demo: read APIs remain unauthenticated.
- Event ingest: `POST /api/v1/events` is configured as an `AWS_IAM` API Gateway route. The application also fails closed outside local mode unless the protected runtime explicitly enables ingest.
- Demo reset: `POST /api/v1/demo/reset` is configured as an `AWS_IAM` route and remains application-disabled in production. It is available only for bounded local demo use by default.
- GitHub evidence producer: a dedicated OIDC role can invoke only the exact ingest method/path on this API.

Status: **IMPLEMENTED AND LOCALLY VALIDATED — NOT YET DEPLOYED OR LIVE VERIFIED**.

## Data protection

- DynamoDB and S3: AWS-managed encryption (SSE-S3 / AWS-owned DDB encryption).
- TLS at CloudFront and API Gateway.
- No secrets in git, Terraform variables, or Angular bundles.
- Prefer IAM + SSM Standard parameters over Secrets Manager until a rotating secret is required (webhook secret is a candidate for SSM SecureString or Secrets Manager; decide at implementation by whether rotation is needed).
- Terraform state: remote S3 backend with encryption and lock, configured when AWS exists. State must not contain webhook secrets as plaintext variables.

## Application

- Strict request size limits
- JSON schema / explicit struct validation
- CORS allowlist (CloudFront origin + localhost in non-prod)
- Security headers on CloudFront / SPA (`Content-Security-Policy` tuned for Angular, `X-Content-Type-Options`, `Referrer-Policy`, `Frame-Options`)
- Sanitized error bodies
- No logging of `Authorization`, webhook signatures, cookies, or env credentials
- Rate limiting at API Gateway
- Browser CORS permits public reads, not production evidence writes

## Supply chain

| Tool | Use | Keep only if real |
| --- | --- | --- |
| Trivy | FS + IaC on PR; image on main; fail CRITICAL | Yes |
| Syft | SBOM for the Go image / binaries | Yes |
| Cosign | Keyless sign/verify on ECR image via GitHub OIDC | Yes if ECR images ship; otherwise omit from README badges |

## Producer authentication

The selected production design is:

1. GitHub Actions requests a short-lived AWS session through GitHub OIDC.
2. The ID-qualified repository and `refs/heads/main` subject assume `cloudops-prod-github-evidence-producer`.
3. The role has only `execute-api:Invoke` for `POST /api/v1/events`.
4. The workflow signs the canonical event request with SigV4 for `execute-api` in `eu-west-1`.
5. API Gateway authorizes the request before the existing Go validation and evidence pipeline run.

A secret is never shipped in the Angular bundle. Every browser user could extract such a secret, so it would not authenticate a trusted producer. CORS is defense in depth for browsers, not authentication.

## Public demo threat model

The demo is intentionally readable. Assume hostile clients will scrape GET APIs and attempt both write routes.

Mitigations:

- API Gateway denies unsigned production ingest and demo-reset requests
- The Go runtime independently denies both writes unless explicitly configured
- Local reset only restores the known synthetic set; no arbitrary writes
- API throttling
- No PII in synthetic data
- Live metadata is this project's own SHAs and deploy times, not user data

## Known limitations

- The IAM routes, producer role and signed workflow are configured locally but are **not yet deployed**.
- The protected release derives the producer role ARN and API origin from the applied Terraform outputs; it stores no AWS credential or anonymous fallback.
- Authentication cannot be called LIVE VERIFIED until an unsigned request is denied and a signed GitHub producer request succeeds in AWS.
- The GitHub plan and application deploy roles retain AWS managed ReadOnlyAccess for compatibility. The production Terraform apply role is separately bounded and scoped to CloudOps resources.
- Cosign remains best-effort in the current workflow; do not describe signing as mandatory enforcement.

See [CV_CLAIMS_MATRIX.md](../CV_CLAIMS_MATRIX.md) for verified claim boundaries.
