# Linux operations

This project uses Linux as a real runtime, not as a résumé keyword and not as an extra EC2 box.

The owner of this repository is learning Linux through the surfaces below. This is not a claim of professional Linux system administration.

## What Linux contributes here

| Surface | Why it matters |
| --- | --- |
| GitHub Actions `ubuntu-latest` | The quality gate, image build, Trivy, and Cypress CI run on Ubuntu |
| OCI images | The deployable Go API/worker is a Linux `amd64` container |
| AWS Lambda `provided.al2023` | Production compute is Amazon Linux, not a Windows process |
| Distroless HTTP image | Local/CI proof of non-root, no shell, SIGTERM |

There is no always-on Linux VM in the architecture.

## Runtime model

- A container shares the host kernel. It is not a second machine.
- Lambda starts the AWS base-image entrypoint, which launches `/var/task/bootstrap`.
- `bootstrap` reads `_HANDLER` (from `CMD` / Terraform `image_config.command`) and `exec`s `api` or `worker`.
- Each Go binary uses `aws-lambda-go` against the Runtime API when `AWS_LAMBDA_RUNTIME_API` is set.
- Local HTTP uses the same graceful-shutdown code as PID 1 in the distroless image.

## Users, groups, permissions

- HTTP image user: `65532:65532` (distroless `nonroot`) — this is the non-root proof surface.
- Lambda image: do **not** set `USER` in the Dockerfile. The `provided.al2023` base entrypoint must run as the image default user; Lambda's execution environment provides the sandbox.
- Writable runtime dir: `HOME=/tmp` on the HTTP image. No `chmod 777`.
- `chmod` changes permission bits. `chown` changes owner. `755` means owner rwx, group rx, other rx.

Why non-root on HTTP: a compromised process should not be able to write the rest of the image or bind privileged ports. For Lambda, isolation is the AWS sandbox; forcing `USER 1000:1000` on the AWS base image breaks the entrypoint contract.

## Processes and signals

- PID 1 in a container receives SIGTERM from `docker stop` / orchestration.
- This API cancels a `signal.NotifyContext` on SIGINT/SIGTERM, then `http.Server.Shutdown`.
- In-flight requests get a bounded wait (`APP_SHUTDOWN_SECONDS`, default 10s).
- Context cancellation is how request work should stop. Do not ignore the request context.

`internal/runtime` tests cancel-driven shutdown. `scripts/linux/graceful-shutdown-test.sh` and `scripts/ci/container-inspect.sh` send SIGTERM to a Linux container in CI.

## Filesystem

- Final images do not contain `go`, `gcc`, or a package manager.
- Distroless has no shell. Debug from the **runner** (`docker image inspect`, `curl`), not `docker exec bash`.
- Secrets are environment / IAM, never baked into layers.

## Networking

- HTTP container listens on `:8080`.
- Production public entry is CloudFront/API Gateway, not an exposed container port on EC2.
- Inspect listening ports on a Linux runner with `ss -lntp` when a process is local to that runner.

## Environment

- `APP_*`, `AWS_REGION`, `DDB_TABLE_NAME`, `EVENT_BUS_NAME`, `S3_RAW_EVENTS_BUCKET`
- Lambda injects `AWS_LAMBDA_RUNTIME_API`. That switches the binary from `ListenAndServe` to the Lambda adapter.
- `printenv` / `env` on the Ubuntu runner are the legitimate way to confirm CI variables. Do not print secrets.

## Logs

- The process writes structured JSON to stdout. CloudWatch captures stdout/stderr from Lambda.
- stdout is the normal log stream. stderr is for diagnostics. This app uses `slog` on stdout.

## Commands actually used

Documented because they appear in scripts, CI, or debugging this project:

| Command | Where |
| --- | --- |
| `uname -a` | CI runner identity |
| `cat /etc/os-release` | CI runner identity |
| `curl` | Health check of the container and API |
| `docker image inspect` | Architecture, user, entrypoint, env |
| `docker run` / `docker stop` / `docker kill --signal=SIGTERM` | Image behaviour and shutdown |
| `printenv` / `env` | Confirm non-secret environment |
| `ls` / `stat` | Only when a shell exists (build stage / Ubuntu runner) |
| `sha256sum` | Optional digest check on the runner |
| `tar` | Image/build context (Docker) |

Not used here, so not claimed: `iptables`, `systemctl`, LVM, SELinux policy writing, kernel builds, fleet-wide admin.

Useful on the Ubuntu runner if a job misbehaves: `ps`, `top`/`htop` if installed, `ss`, `df`, `du`, `grep`, `sed`, `awk`, `find`, `dig`/`nslookup`.

## Container vs VM vs Lambda

- **VM**: hardware isolation, own kernel, billed while running.
- **Container**: isolated processes + namespaces + cgroups, shared kernel.
- **Lambda container**: AWS starts your Linux image per invoke (and reuses warm instances). You do not SSH in.

Namespaces isolate process/network/mount views. Cgroups limit CPU/memory. This project relies on Docker/Lambda for that; it does not configure cgroups by hand.

## Interview questions (honest answers)

**Why run as non-root?** Limit blast radius if the process is exploited. The HTTP image sets USER 65532. The Lambda image relies on the AWS sandbox and must not override the base-image user.

**What is PID 1?** The first process in the container. It must reap children and handle SIGTERM. Our API binary *is* PID 1 in the HTTP image.

**What happens on SIGTERM?** The process should stop accepting connections and finish in-flight work, then exit. `docker stop` sends SIGTERM, then SIGKILL after the grace period.

**chmod vs chown?** Permission bits vs ownership.

**What does 755 mean?** `rwxr-xr-x`.

**How do you inspect listening ports?** `ss -lntp` on Linux. This project checks HTTP with `curl` instead of assuming a shell inside distroless.

**How do environment variables reach a process?** The parent (Docker/Lambda/systemd) puts them in the process environment. Go reads `os.Getenv`.

**What is stdout/stderr?** The two standard streams. Operators collect them as logs.

**Container vs VM?** Shared kernel vs own kernel.

**Namespace/cgroup?** Isolation and resource limits. Know the idea; this repo does not author custom cgroup config.

## What was NOT used

- A standing Ubuntu EC2 instance
- SSH hardening playbooks
- Kubernetes nodes
- Local Docker on the Windows development machine (Docker is not installed there). Linux image proof is CI-on-Ubuntu.

## Claim status

Linux is **IMPLEMENTED** in Dockerfiles, scripts, and GitHub Actions. It becomes **TESTED** when the Ubuntu `linux-container` job is green. It becomes **LIVE VERIFIED** only after the Lambda container serves production traffic.
