#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

PORT="${PF_E2E_PORT:-8098}"
BASE="http://127.0.0.1:${PORT}"
BIN="${ROOT}/bin/platformforge"
PLAYWRIGHT_VERSION="${PF_PLAYWRIGHT_VERSION:-1.51.0}"
PLAYWRIGHT_IMAGE="${PF_PLAYWRIGHT_IMAGE:-mcr.microsoft.com/playwright:v${PLAYWRIGHT_VERSION}-jammy}"

if [[ ! -x "$BIN" ]]; then
  make build
fi

bash scripts/install-e2e-deps.sh

"$BIN" serve --addr "127.0.0.1:${PORT}" &
PID=$!
cleanup() { kill "$PID" 2>/dev/null || true; }
trap cleanup EXIT

for _ in $(seq 1 30); do
  curl -sf "${BASE}/api/health" >/dev/null && break
  sleep 1
done
curl -sf "${BASE}/api/health" | grep -q ok

run_local() {
  cd tests/e2e
  npm install
  export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=0
  npx playwright install chromium
  PF_BASE_URL="$BASE" npx playwright test
}

run_docker() {
  echo "[e2e] running Playwright ${PLAYWRIGHT_VERSION} in Docker (no host browser deps required)"
  docker run --rm \
    --network host \
    -e PF_BASE_URL="${BASE}" \
    -e PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1 \
    -v "${ROOT}/tests/e2e:/work" \
    -w /work \
    "$PLAYWRIGHT_IMAGE" \
    bash -lc "npm install && npx playwright@${PLAYWRIGHT_VERSION} test"
}

if ldconfig -p 2>/dev/null | grep -q libnspr4; then
  run_local
else
  run_docker
fi
