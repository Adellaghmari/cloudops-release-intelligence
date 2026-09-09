# Phase 16 hardening design

Phase 16A prepares; Phase 16B executes reviewed applies and state migration.

## CloudFront TLS truth

Hostname `*.cloudfront.net` uses `CloudFrontDefaultCertificate=true`.
AWS reports viewer `MinimumProtocolVersion=TLSv1` for that certificate.
Terraform must match AWS (`TLSv1`), not fight perpetual drift toward `TLSv1.2_2021`.

Optional polish (user-owned DNS required):

1. ACM certificate in **us-east-1** (CloudFront requirement)
2. DNS validation on the user domain
3. CloudFront alias + ACM ARN
4. `ssl_support_method = sni-only`
5. `minimum_protocol_version = TLSv1.2_2021` (or newer supported policy)

Do not purchase a domain for completion.

## API Gateway X-Ray boundary

Product uses **HTTP API** (`apigatewayv2`), not REST API.

| Surface | Status |
| --- | --- |
| Lambda API Active X-Ray | LIVE VERIFIED |
| Lambda worker Active X-Ray | LIVE VERIFIED |
| HTTP API standalone X-Ray segment | **Not supported** by HTTP API |

This is a cost/simplicity tradeoff, not a defect. Do not migrate to REST API solely for a CV keyword.
DynamoDB subsegments are optional instrumentation; not required for Project Complete.

## GitHub Terraform delivery (after remote state)

| Trigger | Actions |
| --- | --- |
| PR / push (safe) | `fmt`, `validate`, OIDC plan role, `terraform plan` (no apply) |
| Production apply | `workflow_dispatch` and/or GitHub Environment `prod` approval |
| Credentials | GitHub OIDC only — never static `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` |

`ENABLE_TERRAFORM_APPLY` stays false until remote state + locking proven.

## IAM notes

- `ecr:GetAuthorizationToken` legitimately requires `Resource:"*"`.
- CloudFront invalidation waiter needs `cloudfront:GetInvalidation` on the distribution ARN.
- Plan role keeps `ReadOnlyAccess` for refresh; bootstrap adds scoped state lock/object permissions.
- Deploy role product permissions are resource-scoped where AWS allows.

## Project Complete gate (Phase 16B)

Required:

- Public frontend + backend healthy
- CI green; AWS runtime healthy; digests intentional
- Remote state migrated; S3 lockfile proven
- GitHub terraform plan proven; controlled apply proven
- Frontend invalidation waiter proven
- IAM hardening reviewed; no unexplained drift
- Truth matrix accurate; security + cost audits complete
- Git clean; docs current

Optional: custom domain / ACM TLS floor.
