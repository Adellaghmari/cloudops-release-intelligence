# Security

Deliberate controls. No security theatre badges.

## Identity and access

- GitHub Actions: OIDC → IAM deploy role (see [CI_CD.md](CI_CD.md)).
- Lambda: execution roles, not embedded keys.
- DynamoDB/S3/EventBridge/SQS: resource-scoped IAM.
- Public demo: read APIs unauthenticated; writes limited to rate-limited demo reset.
- Ingest: GitHub HMAC and/or IAM-only route, not open to the internet without a secret.

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

## Supply chain

| Tool | Use | Keep only if real |
| --- | --- | --- |
| Trivy | FS + IaC on PR; image on main; fail CRITICAL | Yes |
| Syft | SBOM for the Go image / binaries | Yes |
| Cosign | Keyless sign/verify on ECR image via GitHub OIDC | Yes if ECR images ship; otherwise omit from README badges |

## Webhooks

HMAC SHA-256 verification. Constant-time compare. Reject missing signatures. Do not echo the payload to logs at debug in prod.

## Public demo threat model

The demo is intentionally public. Assume hostile clients will call `POST /api/v1/demo/reset` and scrape APIs.

Mitigations:

- Reset only restores the known synthetic set; no arbitrary writes
- Throttle
- No PII in synthetic data
- Live metadata is this project's own SHAs and deploy times, not user data

## Known limitations (until implemented)

Everything above is specified, not live. See [CV_CLAIMS_MATRIX.md](../CV_CLAIMS_MATRIX.md).
