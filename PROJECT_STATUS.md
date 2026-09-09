# Project status

**Project:** CloudOps Release Intelligence  
**Phase:** 16A COMPLETE (hardening prepared; no new applies) — stop before 16B  
**Status:** Public demo LIVE; Phase 16A plans ready; remote state NOT migrated  
**Complete:** No (Phase 16B execution remains)

## LIVE VERIFIED baseline (unchanged digest)

- Frontend `https://d34fwrlm14h6js.cloudfront.net` HTTP 200
- Backend `https://8kci5uht3d.execute-api.eu-west-1.amazonaws.com` `/health` `/ready` 200
- Lambdas Active on `sha256:ef3778d5af9e80d155610d5ffb0a889e509a4ba3da3fee2ac6878c6c5287ade5`
- Operator `adel-admin` / `eu-west-1` / account `912415493331`

## Phase 16A findings (prepared, not applied)

### CloudFront TLS truth
- Default `*.cloudfront.net` cert reports `MinimumProtocolVersion=TLSv1`
- Terraform now matches AWS (`TLSv1`); stops impossible `TLSv1.2_2021` drift
- Custom domain + ACM in **us-east-1** remains optional polish (user-owned DNS)

### DLQ newest poison closed
- MessageId `149c73ef-9a9a-4996-857e-302be740df4f`
- Body `{poison:true,note:synthetic-dlq-proof-20260909,not_an_envelope:1}`
- Landed on DLQ with ApproximateReceiveCount ≥4 (observed 4 then 5 after peeks)
- Alarm `cloudops-prod-analysis-dlq` State=ALARM
- Labeled synthetic poisons deleted after proof (incl. historical `not-json`); DLQ empty

### API Gateway X-Ray boundary
- HTTP API chosen intentionally; no standalone APIGW X-Ray segment (not a defect)
- Lambda API + worker X-Ray remain LIVE VERIFIED

## Phase 16A.1 correction (prepared, not applied)

- Bootstrap owns **only** the S3 state foundation (no GitHub IAM attachments).
- Main owns GitHub remote-state IAM (plan read+lock; apply get/put state + lock).
- Plan OIDC trust narrowed to `main` + `pull_request` (no `repo:*`).
- New plans: `infra/bootstrap/tfplan-bootstrap-16a1`, `infra/tfplan-hardening-16a1` (informational; discard after migration).
- Migration still uses local `adel-admin` first; GitHub apply remains DISABLED.

### GitHub Terraform apply
- Remains **DISABLED** (local product state; draft `terraform.yml` gated on `TERRAFORM_REMOTE_STATE_READY`)

## Phase 16B execution order (recommended)

1. Apply bootstrap plan → verify state bucket security  
2. Backup local `terraform.tfstate`  
3. Switch main `backend "s3"` + `use_lockfile=true` → `terraform init -migrate-state`  
4. Confirm `plan` 0/0/0 (or only reviewed remainder)  
5. Apply main hardening IAM plan (`tfplan-hardening-16a` or fresh)  
6. Prove frontend invalidation waiter  
7. Enable gated GitHub plan; then controlled `workflow_dispatch` apply  
8. Optional: custom domain/ACM if user provides DNS  

## Cost / security snapshot (16A audit)

- Budget `cloudops-prod-monthly` present  
- NAT/EC2/EKS/RDS/EIP/VPC endpoints = 0  
- Web + raw S3 private (BPA); CloudFront OAC; ECR private; DynamoDB ACTIVE  
