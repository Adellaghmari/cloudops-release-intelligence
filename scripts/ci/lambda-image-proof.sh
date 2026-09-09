#!/usr/bin/env bash
# Prove the Lambda OCI image satisfies the provided.al2023 bootstrap contract
# and can start under the AWS Lambda Runtime Interface Emulator (RIE).
set -euo pipefail

IMAGE="${1:?lambda image tag required}"

echo "== runner Linux =="
uname -a
cat /etc/os-release

echo "== image metadata =="
docker image inspect "$IMAGE" --format 'architecture={{.Architecture}} os={{.Os}} user={{.Config.User}} entrypoint={{json .Config.Entrypoint}} cmd={{json .Config.Cmd}}'
ARCH="$(docker image inspect "$IMAGE" --format '{{.Architecture}}')"
OS="$(docker image inspect "$IMAGE" --format '{{.Os}}')"
ENTRY="$(docker image inspect "$IMAGE" --format '{{json .Config.Entrypoint}}')"
CMD="$(docker image inspect "$IMAGE" --format '{{json .Config.Cmd}}')"

test "$ARCH" = "amd64"
test "$OS" = "linux"
echo "$ENTRY" | grep -q 'lambda-entrypoint'
echo "$CMD" | grep -q 'api'

echo "== filesystem bootstrap contract =="
cid="$(docker create --name "cloudops-lambda-fs-$$" "$IMAGE")"
cleanup_fs() { docker rm -f "$cid" >/dev/null 2>&1 || true; }
trap cleanup_fs EXIT

docker cp "${cid}:/var/task/bootstrap" /tmp/cloudops-bootstrap
docker cp "${cid}:/var/task/api" /tmp/cloudops-api-bin
docker cp "${cid}:/var/task/worker" /tmp/cloudops-worker-bin
docker rm -f "$cid" >/dev/null
trap - EXIT

test -s /tmp/cloudops-bootstrap
test -s /tmp/cloudops-api-bin
test -s /tmp/cloudops-worker-bin
test -x /tmp/cloudops-bootstrap
test -x /tmp/cloudops-api-bin
test -x /tmp/cloudops-worker-bin

# ELF x86-64 for the Go binaries (linux/amd64).
file /tmp/cloudops-api-bin | grep -Eiq 'ELF.*(x86-64|amd64|Intel 80386)'
file /tmp/cloudops-worker-bin | grep -Eiq 'ELF.*(x86-64|amd64|Intel 80386)'
head -n 1 /tmp/cloudops-bootstrap | grep -q '^#!/bin/sh'

echo "== no embedded secrets in task root names =="
docker create --name "cloudops-lambda-ls-$$" "$IMAGE" >/dev/null
docker export "cloudops-lambda-ls-$$" | tar -t | tee /tmp/cloudops-lambda-paths.txt >/dev/null
docker rm -f "cloudops-lambda-ls-$$" >/dev/null
if grep -Eiq '(^|/)(\.env|credentials\.json|\.aws/|id_rsa|terraform\.tfvars)(/|$)' /tmp/cloudops-lambda-paths.txt; then
  echo "forbidden secret-like paths found in image" >&2
  exit 1
fi

echo "== RIE worker invocation (empty SQS batch) =="
rie_cid="$(docker run -d --rm --name "cloudops-rie-$$" -p 9001:8080 \
  -e APP_ENV=local \
  -e APP_STORE=memory \
  -e APP_SEED_LOCAL=false \
  -e APP_LOG_LEVEL=info \
  "$IMAGE" worker)"
cleanup_rie() { docker rm -f "$rie_cid" >/dev/null 2>&1 || true; }
trap cleanup_rie EXIT

ok=0
for _ in $(seq 1 30); do
  if curl -fsS -XPOST "http://127.0.0.1:9001/2015-03-31/functions/function/invocations" \
    -H 'content-type: application/json' \
    -d '{"Records":[]}' -o /tmp/cloudops-rie-out.json; then
    ok=1
    break
  fi
  sleep 1
done
if [ "$ok" != "1" ]; then
  echo "RIE worker invoke failed; container logs:" >&2
  docker logs "$rie_cid" >&2 || true
  exit 1
fi
echo "rie_worker_response=$(cat /tmp/cloudops-rie-out.json)"
# Empty batch should not return a Lambda runtime error object.
if grep -Eqi '"error(Type|Message)"' /tmp/cloudops-rie-out.json; then
  echo "unexpected Lambda error payload from RIE worker invoke" >&2
  cat /tmp/cloudops-rie-out.json >&2
  exit 1
fi
docker rm -f "$rie_cid" >/dev/null
trap - EXIT

echo "lambda image bootstrap + RIE proof OK"
