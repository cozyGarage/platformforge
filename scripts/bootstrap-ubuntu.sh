#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

log() { printf '[bootstrap] %s\n' "$*"; }
need_cmd() { command -v "$1" >/dev/null 2>&1; }

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "Run this script inside Ubuntu WSL2." >&2
  exit 1
fi

log "Ensuring base packages"
if need_cmd curl && need_cmd git && need_cmd make && need_cmd sqlite3; then
  log "Base packages already present"
else
  sudo apt-get update -qq
  sudo apt-get install -y -qq \
    build-essential curl git make sqlite3 jq ca-certificates gnupg lsb-release
fi

GO_VERSION="1.25.12"
install_go=false
if ! need_cmd go; then
  install_go=true
else
  current="$(go env GOVERSION 2>/dev/null | sed 's/^go//' || true)"
  if [[ -z "$current" || "$(printf '%s\n%s\n' "$GO_VERSION" "$current" | sort -V | head -1)" != "$GO_VERSION" ]]; then
    install_go=true
  fi
fi
if [[ "$install_go" == true ]]; then
  log "Installing Go ${GO_VERSION}"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tgz
  rm /tmp/go.tgz
  export PATH="/usr/local/go/bin:${PATH}"
fi

NODE_MAJOR_REQUIRED=22
install_node=false
if ! need_cmd node; then
  install_node=true
else
  node_major="$(node -p "process.versions.node.split('.')[0]" 2>/dev/null || echo 0)"
  if [[ "${node_major}" -lt "${NODE_MAJOR_REQUIRED}" ]]; then
    install_node=true
  fi
fi
if [[ "$install_node" == true ]]; then
  log "Installing Node.js ${NODE_MAJOR_REQUIRED}"
  curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR_REQUIRED}.x" | sudo -E bash -
  sudo apt-get install -y -qq nodejs
fi

if ! need_cmd docker; then
  log "Docker not found; install Docker Engine before continuing"
  exit 1
fi

if ! groups "$USER" | grep -q docker; then
  log "Adding $USER to docker group (re-login required)"
  sudo usermod -aG docker "$USER" || true
fi

export PATH="/usr/local/go/bin:${HOME}/.local/bin:${PATH}"
export GOPATH="${GOPATH:-$HOME/go}"
mkdir -p "${HOME}/.local/bin"

log "Go: $(go version)"
log "Node: $(node --version)"
log "npm: $(npm --version)"
log "Docker: $(docker --version)"

log "Fetching Go modules"
go mod download

log "Installing web dependencies"
(cd web && npm ci)

log "Building PlatformForge"
make build

if ! need_cmd k3d; then
  log "Installing k3d to ~/.local/bin"
  curl -fsSL https://github.com/k3d-io/k3d/releases/download/v5.8.3/k3d-linux-amd64 -o "${HOME}/.local/bin/k3d"
  chmod +x "${HOME}/.local/bin/k3d"
fi

if ! need_cmd kubectl; then
  log "Installing kubectl to ~/.local/bin"
  KUBECTL_VERSION="$(curl -fsSL https://dl.k8s.io/release/stable.txt)"
  curl -fsSL "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" -o "${HOME}/.local/bin/kubectl"
  chmod +x "${HOME}/.local/bin/kubectl"
fi

log "k3d: $(k3d version 2>/dev/null | head -1 || echo present)"
log "kubectl: $(kubectl version --client 2>/dev/null | head -1 || echo present)"

if [[ -f scripts/install-e2e-deps.sh ]]; then
  bash scripts/install-e2e-deps.sh || true
fi

log "Bootstrap complete. Run: ./bin/platformforge serve"
