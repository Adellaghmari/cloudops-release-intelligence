# Cost

Portfolio constraint: this must stay cheap at idle and cheap under recruiter traffic.

## Target

**Expected idle + light public demo: about $2–8 USD / month** in `eu-west-1`, assuming 14-day log retention and no NAT/EKS/RDS.

This is an estimate, not a quote. It will be replaced with actual AWS Cost Explorer figures after go-live.

## Intended services and why they are cheap here

| Service | Pricing posture | Idle expectation |
| --- | --- | --- |
| CloudFront | HTTPS + small static site | Low; pennies to a couple of dollars |
| S3 | Static origin + raw JSON objects | Pennies |
| API Gateway HTTP API | Pay per request; generous free tier | Near zero idle |
| Lambda | Pay per invoke; free tier | Near zero idle |
| DynamoDB on-demand | Pay per request; no capacity to waste | Pennies at demo scale |
| EventBridge custom bus | $1 / million events | Near zero |
| SQS + DLQ | Free tier then tiny | Near zero |
| ECR | Storage of one small Go image | Pennies |
| CloudWatch Logs | **Watch this** | Keep retention 14 days |
| X-Ray | First 100k traces/month free | Recruiter traffic should fit |
| AWS Budgets | First two budgets free | $0 |
| SSM Parameter Store standard | Free for standard params | $0 if used |

## Explicitly excluded (material recurring cost)

Do not provision these unless the user explicitly approves a new cost model:

| Service | Why excluded |
| --- | --- |
| NAT Gateway | ~$32/month idle plus data. Avoided by not putting Lambda in a VPC |
| EKS | Control plane ~$72/month plus nodes |
| RDS / Aurora | Always-on database cost; DynamoDB replaces it |
| ALB in front of always-on compute | Unnecessary with API Gateway |
| AWS WAF (full) | Can add up; start with API Gateway throttling and CloudFront defaults |
| Secrets Manager | $0.40/secret/month. Prefer IAM roles and SSM unless a rotating secret is truly required |
| Multi-AZ always-on anything | No |
| Multi-region active-active | No |
| VPC endpoints "for completeness" | Interface endpoints have hourly cost |

If a future design seems to require NAT, stop and redesign (usually: keep Lambda in AWS-owned network, not a private subnet).

## Budget protection

When Terraform is applied:

1. AWS Budget `cloudops-prod-monthly` with **$5** and **$10** ACTUAL alerts (limit $10). First bootstrap applied 2026-09-09.
2. CloudWatch log groups: retention **14 days**.
3. DynamoDB: `PAY_PER_REQUEST` only. No provisioned capacity.
4. Lambda: memory sized for cold-start vs cost after first traces exist; start at 512 MB API / 512 MB worker.
5. S3 lifecycle: raw events expire after **90 days**.
6. X-Ray sampling: conservative after free-tier headroom is understood.

Alert destination: set `budget_notification_email` via gitignored `infra/terraform.tfvars` or `TF_VAR_budget_notification_email`. Do not commit the address. No new paid notification bus.

## Demo-reset cost control

`POST /api/v1/demo/reset` rewrites a bounded synthetic dataset (fixed service count, fixed scenario count). It must not generate unbounded time-series. Health snapshots are window aggregates (one baseline + one post window per deployment), not per-second points.

API Gateway throttle: conservative burst for public GETs; tighter quota for demo reset (for example a handful per minute).

## Cost interview sentence

> The architecture is serverless on purpose. The expensive AWS defaults — NAT, EKS, RDS — were rejected because they do not serve the traffic shape. The remaining risk is log retention, which is capped at 14 days, plus a $5/$10 budget alarm.
