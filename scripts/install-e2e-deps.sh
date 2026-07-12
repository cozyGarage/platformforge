#!/usr/bin/env bash
set -euo pipefail

has_playwright_deps() {
  ldconfig -p 2>/dev/null | grep -q libnspr4 || return 1
  return 0
}

if has_playwright_deps; then
  echo "[e2e-deps] Playwright system libraries present"
  exit 0
fi

if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  echo "[e2e-deps] System libs missing; e2e will use Playwright Docker image"
  exit 0
fi

echo "[e2e-deps] Install browser libraries manually:"
echo "  sudo bash scripts/install-e2e-deps.sh"
echo "  or ensure Docker is running for containerized Playwright"
exit 1
