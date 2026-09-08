#!/usr/bin/env bash
# Inspect the local HTTP Linux image. Distroless has no shell; use docker inspect
# plus an HTTP health check instead of docker exec.
set -euo pipefail

IMAGE="${1:?image tag required}"

echo "== runner Linux =="
uname -a
cat /etc/os-release

echo "== image metadata =="
docker image inspect "$IMAGE" --format 'architecture={{.Architecture}} os={{.Os}} user={{.Config.User}} entrypoint={{json .Config.Entrypoint}}'
docker image inspect "$IMAGE" --format 'env={{json .Config.Env}}'

USER_VAL="$(docker image inspect "$IMAGE" --format '{{.Config.User}}')"
if [ "$USER_VAL" = "0" ] || [ "$USER_VAL" = "root" ]; then
  echo "production HTTP image must not run as root" >&2
  exit 1
fi

echo "== health + SIGTERM =="
cid="$(docker run -d -p 18080:8080 -e APP_ENV=local -e APP_HTTP_ADDR=:8080 -e APP_SEED_LOCAL=true "$IMAGE")"
cleanup() { docker rm -f "$cid" >/dev/null 2>&1 || true; }
trap cleanup EXIT

for _ in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:18080/api/v1/health" >/dev/null; then
    break
  fi
  sleep 1
done
curl -fsS "http://127.0.0.1:18080/api/v1/health"
echo
docker stop -t 8 "$cid" >/dev/null
echo "container accepted SIGTERM and stopped"
