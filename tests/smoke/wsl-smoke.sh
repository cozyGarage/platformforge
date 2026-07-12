#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

PORT="${PF_SMOKE_PORT:-8099}"
BASE="http://127.0.0.1:${PORT}"
BIN="${ROOT}/bin/platformforge"
LAB="${PF_SMOKE_LAB:-linux-navigation}"

log() { printf '[smoke] %s\n' "$*"; }

if [[ ! -x "$BIN" ]]; then
  log "building binary"
  make build
fi

log "starting server on ${BASE}"
"$BIN" serve --addr "127.0.0.1:${PORT}" &
PID=$!
cleanup() { kill "$PID" 2>/dev/null || true; }
trap cleanup EXIT

for _ in $(seq 1 30); do
  curl -sf "${BASE}/api/health" >/dev/null && break
  sleep 1
done

curl -sf "${BASE}/api/health" | grep -q ok
curl -sf "${BASE}/api/labs" | grep -q "$LAB"
curl -sf "${BASE}/api/labs/${LAB}/status" | grep -q '"running":false'

log "starting lab ${LAB}"
curl -sf -X POST "${BASE}/api/labs/${LAB}/start" >/dev/null
sleep 3
curl -sf "${BASE}/api/labs/${LAB}/status" | grep -q '"running":true'

log "validating lab (expect failures before learner work — smoke checks API path)"
curl -sf -X POST "${BASE}/api/labs/${LAB}/validate" | grep -q checks

log "resetting lab"
curl -sf -X POST "${BASE}/api/labs/${LAB}/reset" >/dev/null
curl -sf -X POST "${BASE}/api/labs/${LAB}/stop" -o /dev/null -w '%{http_code}' | grep -q 204
curl -sf "${BASE}/api/labs/${LAB}/status" | grep -q '"running":false'

log "smoke passed"
