#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

PORT="${PF_E2E_PORT:-8098}"
BASE="http://127.0.0.1:${PORT}"
BIN="${ROOT}/bin/platformforge"

if [[ ! -x "$BIN" ]]; then
  make build
fi

"$BIN" serve --addr "127.0.0.1:${PORT}" &
PID=$!
cleanup() { kill "$PID" 2>/dev/null || true; }
trap cleanup EXIT

for _ in $(seq 1 30); do
  curl -sf "${BASE}/api/health" >/dev/null && break
  sleep 1
done

cd tests/e2e
npm install
npx playwright install chromium
if command -v apt-get >/dev/null 2>&1; then
  npx playwright install-deps chromium 2>/dev/null || echo "[e2e] install browser deps: npx playwright install-deps chromium"
fi
PF_BASE_URL="$BASE" npx playwright test
