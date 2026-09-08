# Interview notes

Portfolio talking points grounded in this design. Do not claim LIVE VERIFIED items until the matrix says so.

## Why Go?

The other flagship project already covers Python/FastAPI. This backend needs strict types, cheap Lambda binaries, and engines that are ordinary functions under `testing`. Go fits an event worker + HTTP API in one module without a language split.

## Why DynamoDB?

Access patterns are key-value and time-sorted lists around `release` and `service`. A relational schema would spend budget on RDS idle cost and would duplicate Project A's PostgreSQL story. Single-table keys are documented in [DYNAMODB_DESIGN.md](DYNAMODB_DESIGN.md).

## Why event-driven? Why EventBridge and SQS?

Ingest must acknowledge quickly. Analysis (graph + OPA + risk + health) is slower and must retry independently. EventBridge is the typed routing bus; SQS is the buffer with visibility timeout and DLQ. EventBridge without SQS would still need somewhere to park failures. SQS without EventBridge would couple producers to the worker queue.

## What if a message is delivered twice?

`event_id` conditional writes + per-kind analysis idempotency. Duplicates are success.

## What if the worker fails?

Three receives, then DLQ, then alarm. No infinite loop. Timeline already has the raw event.

## Why a DLQ?

Poison messages (bugs, unexpected payload after schema drift) must not block the queue. DLQ makes that visible.

## How does risk work? Why is it not probability?

Additive weighted signals with published caps. No dataset of thousands of failures was used to calibrate probabilities. Calling it a probability would be dishonest. See [RELEASE_RISK_ENGINE.md](RELEASE_RISK_ENGINE.md).

## How does health comparison avoid false correlation?

Warmup gap, minimum sample, absolute floors, overlapping-deploy suppression, baseline-already-bad → UNLIKELY, and language that refuses causation. See [HEALTH_COMPARISON.md](HEALTH_COMPARISON.md).

## How does the graph work? Cycles?

BFS on reverse edges for dependents. DFS coloring for cycle listing. `seen` prevents revisiting. Reachability ≠ outage.

## Why policy as code? How does OPA run?

Rego is versioned in-repo. OPA is a library in-process. Evaluations store rule ID, version, input excerpt. Fail closed on engine error.

## How is rollback readiness calculated?

Checklist of metadata prerequisites. PARTIAL vs NOT_READY is explicit. No auto rollback.

## Terraform state?

Remote S3 + lock when AWS exists. Local state is not for prod. Secrets do not belong in `tfvars`.

## Why GitHub OIDC rather than static keys?

Keys leak and live forever unless rotated. OIDC issues short-lived credentials bound to repo and branch.

## IAM least privilege?

Separate plan vs apply roles. Lambda roles scoped to table, bucket, bus, queue. No `*:*`.

## Trace a production release to git?

Lambda digest → ECR tag SHA → GitHub commit. Frontend prefix `web/<sha>/`. Dogfood event in the timeline.

## How is production observed?

JSON logs, CloudWatch metrics, X-Ray, DLQ alarm. Not Application Insights.

## Cost compromises?

No NAT, no EKS, no RDS, one region, 14-day logs, optional no staging. That is a portfolio choice, not how a bank would isolate accounts.

## What this project does not prove

Kubernetes expertise, LLM systems, Azure, on-call in a real company, scientific causality.
