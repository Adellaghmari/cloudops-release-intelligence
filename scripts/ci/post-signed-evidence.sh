#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: post-signed-evidence.sh <api-base-url> <event-json-file>" >&2
  exit 2
fi

api_base="${1%/}"
payload_file="$2"

: "${AWS_REGION:?AWS_REGION is required}"
: "${AWS_ACCESS_KEY_ID:?OIDC-provided AWS_ACCESS_KEY_ID is required}"
: "${AWS_SECRET_ACCESS_KEY:?OIDC-provided AWS_SECRET_ACCESS_KEY is required}"
: "${AWS_SESSION_TOKEN:?OIDC-provided AWS_SESSION_TOKEN is required}"

if [[ ! "$api_base" =~ ^https://[^/?#]+$ ]]; then
  echo "API base URL must be an HTTPS origin without a path, query, or fragment" >&2
  exit 2
fi
if [[ ! -s "$payload_file" ]]; then
  echo "event payload file is missing or empty" >&2
  exit 2
fi

headers_file="$(mktemp)"
response_file="$(mktemp)"
trap 'rm -f "$headers_file" "$response_file"' EXIT

curl --fail-with-body --silent --show-error \
  --request POST \
  --aws-sigv4 "aws:amz:${AWS_REGION}:execute-api" \
  --user "${AWS_ACCESS_KEY_ID}:${AWS_SECRET_ACCESS_KEY}" \
  --header "x-amz-security-token: ${AWS_SESSION_TOKEN}" \
  --header "content-type: application/json" \
  --header "x-correlation-id: gha-${GITHUB_RUN_ID:-local}" \
  --dump-header "$headers_file" \
  --output "$response_file" \
  --data-binary "@${payload_file}" \
  "${api_base}/api/v1/events"

request_id="$(awk 'BEGIN{IGNORECASE=1} /^x-request-id:/ {gsub(/\r/,"",$2); print $2}' "$headers_file" | tail -n 1)"
echo "Authenticated event ingestion accepted${request_id:+; request_id=${request_id}}"
