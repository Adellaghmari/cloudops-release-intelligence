#!/usr/bin/env bash
# Sends SIGTERM to a Linux HTTP container and expects a clean stop.
set -euo pipefail

IMAGE="${1:?image tag required}"
cid="$(docker run -d -e APP_ENV=local -e APP_HTTP_ADDR=:8080 -e APP_SEED_LOCAL=true -p 18081:8080 "$IMAGE")"
cleanup() { docker rm -f "$cid" >/dev/null 2>&1 || true; }
trap cleanup EXIT

for _ in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:18081/api/v1/health" >/dev/null; then
    break
  fi
  sleep 1
done

docker kill --signal=SIGTERM "$cid" >/dev/null
for _ in $(seq 1 20); do
  if [ "$(docker inspect -f '{{.State.Running}}' "$cid")" = "false" ]; then
    echo "SIGTERM produced a clean container exit"
    exit 0
  fi
  sleep 0.5
done
echo "container did not stop after SIGTERM" >&2
exit 1
