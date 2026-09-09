# Cost checkpoint — before first `terraform apply`

This file is the stop-the-line review required before any AWS resource is created.

**Status: FIRST BOOTSTRAP APPLIED (2026-09-09).** Foundation resources exist in `eu-west-1`. Lambda, API Gateway, and the public demo are still not created.

This file remains the cost checkpoint. It is no longer a “do not apply” gate for the first plan.

## Resource list (intended first apply)

Always created (no always-on compute):

| Resource | Count | Idle nature |
| --- | --- | --- |
| DynamoDB table `PAY_PER_REQUEST` + 2 GSIs | 1 | $0 with no traffic |
| S3 web bucket + OAC + encryption | 1 | Pennies for empty/small site |
| S3 raw evidence bucket + 90-day lifecycle | 1 | Pennies |
| ECR repository | 1 | Pennies for one small Go image |
| EventBridge custom bus + 1 rule | 1 | Near zero |
| SQS analysis queue + DLQ | 2 | Near zero |
| CloudWatch log groups (14-day retention) | 3 | Watch this |
| CloudWatch DLQ alarm | 1 | Free tier / pennies |
| CloudFront distribution | 1 | Low; HTTPS default cert |
| AWS Budget $10 with optional $5/$10 email | 1 | First two budgets free |
| IAM roles (API, worker, GitHub deploy, GitHub plan) | 4 | $0 |
| GitHub OIDC provider | 0 or 1 | $0 |

Created only after an immutable image digest exists (`api_image_uri`):

| Resource | Count |
| --- | --- |
| Lambda API (container image, 512 MB) | 1 |
| Lambda worker (same image, different CMD) | 1 |
| Lambda SQS event source mapping | 1 |
| API Gateway HTTP API + default stage | 1 |

Explicitly **not** in this plan:

- NAT Gateway
- EKS
- RDS / Aurora
- Always-on EC2
- VPC-attached Lambda
- ALB
- WAF
- Secrets Manager
- Multi-region

## Estimated idle cost

Target remains about **$2–8 USD / month**.

Realistic idle with an empty or near-empty demo:

- CloudFront + S3: about $0–2
- ECR storage: pennies
- CloudWatch Logs: about $0–2 if retention stays 14 days
- Everything else: effectively $0 at idle

This is an estimate, not a quote.

## Expected recruiter / demo usage

A few dozen page loads and API reads per day:

- API Gateway HTTP API + Lambda: free-tier / cents
- DynamoDB on-demand: cents
- EventBridge + SQS: cents
- X-Ray: should stay inside the first 100k traces

Demo reset rewrites a bounded Northstar catalog. It does not write unbounded time series.

## Surprise-cost services

| Service | Why it can surprise | Control |
| --- | --- | --- |
| CloudWatch Logs | Retention and verbose logs | 14-day retention, JSON access logs only |
| CloudFront | Unexpected traffic | PriceClass_100, no custom WAF |
| X-Ray | High sampling later | Active tracing on two Lambdas only |
| S3 request volume | Reset + evidence writes | 90-day raw expiry, no public listing |

## Budget

- Warning: **$5** actual-spend notification (email)
- Stronger warning: **$10** actual-spend notification (email); budget limit is also $10
- Subscriber email is a **required Terraform input**, not committed:

```powershell
$env:TF_VAR_budget_notification_email = "you@example.com"
```

Or copy `infra/terraform.tfvars.example` to `infra/terraform.tfvars` (gitignored) and set `budget_notification_email`.


## Apply decision

Estimated recurring idle cost stays inside the original low single-digit USD target.

**First bootstrap:** applied 2026-09-09 (`34 added, 0 changed, 0 destroyed`). Post-apply inspection found no NAT Gateway, EKS, RDS, EC2, Elastic IP, VPC endpoint, Lambda function, or API Gateway.

**Do not run GitHub Actions terraform apply** while state is local (`infra/terraform.tfstate`, gitignored). That would risk a second stack.

**Second bootstrap (not started):** push an immutable ECR image digest, then apply with `api_image_uri` set. That creates Lambda + HTTP API. Do not start that until explicitly approved.
