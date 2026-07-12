# Ubuntu WSL2 development

## Host configuration

Use Windows `%USERPROFILE%\.wslconfig` for VM-wide settings:

```ini
[wsl2]
memory=6GB
processors=3
swap=2GB
localhostForwarding=true

[experimental]
autoMemoryReclaim=gradual
```

Use `/etc/wsl.conf` inside Ubuntu for distro settings:

```ini
[boot]
systemd=true

[user]
default=home
```

After changing either file, run `wsl --shutdown` from PowerShell, reopen Ubuntu, then run:

```bash
cd ~/Projects/platformforge
./scripts/bootstrap-ubuntu.sh
./scripts/doctor.sh
```

## Development loop

Run `go run ./cmd/platformforge serve` for the API and `npm run dev --prefix web` for Vite. Vite proxies HTTP and WebSocket API traffic to port 8080.

Production assets are copied into `internal/ui/dist` and embedded in the binary by `make build`.

## Common WSL problems

- Keep the repository in the Linux filesystem, not `/mnt/c`, for reliable permissions and filesystem performance.
- If Docker is unreachable, check `systemctl status docker` and group membership with `id`.
- If localhost is unavailable after a configuration change, run `wsl --shutdown` and restart Ubuntu.
- Mirrored networking and DNS tunneling require newer Windows builds; on Windows 10 19045 use NAT mode (default).

# Development guide

PlatformForge is a Go + React monorepo intended to run entirely inside Ubuntu WSL2.

## Prerequisites

- Ubuntu WSL2 with systemd enabled in `/etc/wsl.conf`
- Docker Engine running in WSL
- Windows `.wslconfig` should only contain `[wsl2]` and `[experimental]` keys; put `[boot] systemd=true` in Ubuntu's `/etc/wsl.conf`

Recommended host allocation (13.9 GB RAM machine):

```ini
[wsl2]
memory=6GB
processors=3
swap=2GB
```

## Bootstrap

```bash
bash scripts/bootstrap-ubuntu.sh
bash scripts/doctor.sh
```

## Build layout

| Path | Purpose |
|------|---------|
| `cmd/platformforge/` | CLI entrypoint |
| `internal/ui/` | Embedded Vite production assets |
| `internal/` | API, lab engine, Docker runtime, progress store |
| `web/` | React + Vite learner UI |
| `content/platform-foundations/` | Markdown lessons + YAML lab manifests |
| `schemas/lab.schema.json` | Lab manifest JSON Schema |

Build pipeline:

1. `cd web && npm run build` produces `web/dist`
2. Copy to `internal/ui/dist/`
3. `go build -o bin/platformforge ./cmd/platformforge`

`make build` runs all steps.

## Local development

Terminal 1 — API:

```bash
make build
./bin/platformforge serve --addr 127.0.0.1:8080
```

Terminal 2 — UI hot reload (optional):

```bash
cd web && npm run dev
```

Vite proxies `/api` and WebSocket traffic to Go.

## Lab authoring

Each lab directory contains:

- `lab.yaml` — manifest validated against `schemas/lab.schema.json`
- `lesson.md` — learner-facing narrative

Supported check types: `command`, `file`, `process`, `port`, `http`, `docker`, `kubernetes`.

## Testing

```bash
make test
make lint
make smoke
```

## Common WSL pitfalls

- Running bootstrap from `/mnt/c/...` instead of `~/Projects/...` — bind mounts and performance suffer; keep the repo in the Linux filesystem.
- Docker socket permission denied — add your user to the `docker` group and restart WSL.
- Invalid `.wslconfig` keys under `[boot]` — move systemd config to `/etc/wsl.conf`.
