# Project status

**Project:** CloudOps Release Intelligence  
**Phase:** 15 COMPLETE (public recruiter demo LIVE) — stop before Phase 16  
**Status:** Public frontend + backend operating together  
**Complete:** No (Phase 16 hardening remains)

## LIVE VERIFIED (account `912415493331`, `eu-west-1`)

### Public recruiter demo (Phase 15)
- Frontend `https://d34fwrlm14h6js.cloudfront.net` HTTP 200 (Angular 21 production SPA)
- SPA → API Gateway `https://8kci5uht3d.execute-api.eu-west-1.amazonaws.com` (prod `apiBase` absolute; no localhost in bundle)
- Private S3 origin `cloudops-prod-web-7be25877` (Block Public Access all true; PolicyStatus IsPublic=false; CloudFront OAC `E3TW0G8ABTQWD8`)
- CloudFront invalidation `IOLP72QIXMZIGX0SAGC7SUH68` completed after first upload
- SPA routing fallback: direct GET `/`, `/services`, `/releases`, `/releases/rel_northstar_payments_demo`, `/architecture`, `/replay`, `/status` → 200 HTML (not S3 XML)
- Browser proof: Operations overview, release detail (risk/health/impact/policy/rollback/timeline), System Status LIVE PROJECT DATA
- CORS from CloudFront origin allowed; `evil.example` gets no `Access-Control-Allow-Origin`
- LIVE vs SYNTHETIC labeling in UI; dogfood row `evt_gha_phase15a1b2c3d4` shows git SHA `9710098e…`, CD run `34379595587`, digest `sha256:ef3778d5…`
- First SPA upload: manual `adel-admin` `aws s3 sync` + invalidation `IOLP72QIXMZIGX0SAGC7SUH68`
- GitHub OIDC frontend deploy LIVE: [cd 34384490644](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34384490644) — OIDC assume + production build + S3 sync + CreateInvalidation; Lambda skipped; digests unchanged `ef3778d5…`. Prior attempt [34384226392](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34384226392) failed only on waiter (`GetInvalidation` missing; waiter removed; IAM expansion deferred to Phase 16)
- CD path filters skip Lambda on frontend/docs/workflow-only pushes; `workflow_dispatch` Lambda rebuild requires explicit `force_lambda` input

### Backend (unchanged digest)
- Lambdas remain `sha256:ef3778d5af9e80d155610d5ffb0a889e509a4ba3da3fee2ac6878c6c5287ade5` (Git `9710098e…`) — not rebuilt for frontend/docs
- Public API health/ready/services/releases + analysis surfaces HTTP 200
- DynamoDB `cloudops-prod-main`; Northstar SYNTHETIC seed
- EventBridge → SQS → worker; EventID idempotency; CloudWatch; X-Ray (prior session)
- DLQ/retry: historical DLQ body `not-json` ReceiveCount=4 (ALARM); newer poison `149c73ef…` first fail observed — full 3×360s wait for that message was not finished in-session
- CI/CD, OIDC, Trivy, Syft, Cosign (prior); CD path filters now skip Lambda image build on frontend/docs-only pushes

### CloudFront TLS apply attempt
- Fresh plan exactly `0 add / 2 change / 0 destroy` (TLS floor desired `TLSv1.2_2021` + S3 web policy refresh) → applied saved `tfplan-cloudfront-tls-1`
- AWS still reports ViewerCertificate `MinimumProtocolVersion=TLSv1` for default `*.cloudfront.net` certificate; Terraform desired state remains `TLSv1.2_2021` → persistent plan drift `0/2/0` (no resource replacement). True TLS 1.2 floor reporting needs custom domain + ACM (Phase 16)

## NOT LIVE VERIFIED / Phase 16 hardening

- Add `cloudfront:GetInvalidation` to deploy role (Terraform IAM review) so CD can wait on invalidations
- Custom domain + ACM so CloudFront can enforce/report TLSv1.2_2021 (default cert reports TLSv1)
- Full in-session wait for newest poison message through all 3 receives into DLQ (~18 min)
- API Gateway / DynamoDB as distinct X-Ray subsegments
- GitHub Terraform apply (local state only — remains DISABLED)
- Remote Terraform backend
- Project C

## Cost posture

Budget `cloudops-prod-monthly` configured. No NAT Gateway, no running EC2, no EKS, no RDS, no Elastic IP, no VPC endpoint observed in this check. Serverless portfolio stack only.
