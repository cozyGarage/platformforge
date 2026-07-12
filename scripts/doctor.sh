#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

pass=0
fail=0
warn=0

ok() { printf '  [OK]   %s\n' "$*"; pass=$((pass + 1)); }
bad() { printf '  [FAIL] %s\n' "$*"; fail=$((fail + 1)); }
note() { printf '  [WARN] %s\n' "$*"; warn=$((warn + 1)); }

check_cmd() {
  if command -v "$1" >/dev/null 2>&1; then ok "$1 found: $($1 --version 2>/dev/null | head -1 || echo present)"; else bad "$1 not found"; fi
}

printf 'PlatformForge doctor\n\n'

if [[ "$(uname -s)" == "Linux" && -f /proc/version ]] && grep -qi microsoft /proc/version; then
  ok "Running inside WSL"
else
  note "Not detected as WSL; local-first labs expect Ubuntu WSL2"
fi

if [[ -f /etc/wsl.conf ]] && grep -q 'systemd=true' /etc/wsl.conf; then
  ok "systemd enabled in /etc/wsl.conf"
else
  bad "systemd=true missing from /etc/wsl.conf"
fi

check_cmd go
check_cmd node
check_cmd npm
check_cmd docker
check_cmd git
check_cmd sqlite3
check_cmd make

if command -v k3d >/dev/null 2>&1; then ok "k3d found: $(k3d version 2>/dev/null | head -1 || echo present)"; else note "k3d missing (required for kubernetes-deploy lab)"; fi
if command -v kubectl >/dev/null 2>&1; then ok "kubectl found: $(kubectl version --client 2>/dev/null | head -1 || echo present)"; else note "kubectl missing (required for kubernetes-deploy lab)"; fi

if docker info >/dev/null 2>&1; then
  ok "Docker daemon reachable"
else
  bad "Docker daemon not reachable (is dockerd running?)"
fi

if [[ -f go.mod ]]; then
  ok "go.mod present"
else
  bad "go.mod missing (run from repo root)"
fi

if [[ -d web/node_modules ]]; then
  ok "web dependencies installed"
else
  note "web/node_modules missing (run scripts/bootstrap-ubuntu.sh)"
fi

if [[ -x bin/platformforge ]]; then
  ok "binary built at bin/platformforge"
else
  note "binary not built yet (run make build)"
fi

printf '\nSummary: %d passed, %d failed, %d warnings\n' "$pass" "$fail" "$warn"
[[ "$fail" -eq 0 ]]
