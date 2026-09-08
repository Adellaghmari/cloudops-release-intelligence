#!/usr/bin/env bash
set -euo pipefail

if [ -n "$(gofmt -l cmd internal)" ]; then
  echo "gofmt required:" >&2
  gofmt -l cmd internal
  exit 1
fi

go vet ./cmd/... ./internal/...
go test ./cmd/... ./internal/...
