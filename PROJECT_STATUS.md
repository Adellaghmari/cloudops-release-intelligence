# Project status

**Project:** CloudOps Release Intelligence
**Phase:** 11–15 (DynamoDB credential-chain fix preparing next image)
**Status:** NOT COMPLETE — bootstrap LIVE; public API blocked until credential-chain image is applied
**Complete:** No

## Current truth (AWS)

LIVE VERIFIED:

- Terraform foundation + compute + runtime-fix apply for digest `sha256:b2569a0452d2be9c1be0e509db7dea7bc0a12d202e714334702cb2de1985c40f`
- Lambda bootstrap / `Runtime.InvalidEntrypoint` resolved on AWS
- Worker empty-SQS invoke + X-Ray trace `1-6aa18d4e-5cd76c662800c7b051f6373b`
- GitHub Actions CI/CD, OIDC, Trivy, Syft, Cosign for `b2569a04…`
- S3, CloudFront (empty SPA), ECR, EventBridge/SQS resources, Budget, CloudWatch groups

NOT LIVE VERIFIED:

- Public API (`/api/v1/*` HTTP 500 while `b2569a04…` remains deployed)
- DynamoDB application persistence (ItemCount 0)
- Northstar AWS seed, EventBridge→SQS→worker domain path, duplicate EventID, DLQ E2E, CORS, successful API X-Ray

## Confirmed blocker (source fixed; image not yet applied)

`internal/repository/dynamo/client.go` previously always injected `local`/`local` static credentials, overriding the Lambda execution-role chain. Fix: static credentials only when `AWS_ENDPOINT_URL` / `ClientOptions.Endpoint` is set (DynamoDB Local). Production uses SDK default credential chain + region.

## Remaining before public Angular

1. Push credential-chain image → review `tfplan-dynamodb-credentials-fix-1` → apply
2. Prove health/ready/services/releases + seed idempotency
3. Prove EventBridge→SQS→worker→DynamoDB, duplicate EventID, DLQ
4. CORS + API X-Ray success
5. Upload Angular (Phase 15)
6. Remote Terraform state before GitHub apply

Phase 16 / Project C not started.
