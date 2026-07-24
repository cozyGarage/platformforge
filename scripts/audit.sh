#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> go mod verify"
go mod verify

echo "==> govulncheck"
if ! command -v govulncheck >/dev/null 2>&1; then
  go install golang.org/x/vuln/cmd/govulncheck@latest
  export PATH="$(go env GOPATH)/bin:${PATH}"
fi
govulncheck ./...

echo "==> npm audit (production)"
npm audit --prefix web --omit=dev

echo "==> npm audit (all)"
npm audit --prefix web

echo "audit passed"
