# PlatformForge

PlatformForge is a local-first platform-engineering school with concise lessons, isolated Docker labs, a browser terminal, deterministic validation, reset/retry, and local progress.

## Quick start on Ubuntu WSL2

```bash
git clone <your-repository-url> ~/Projects/platformforge
cd ~/Projects/platformforge
./scripts/bootstrap-ubuntu.sh
./bin/platformforge serve
```

Open http://127.0.0.1:8080.

## CLI

```bash
platformforge doctor
platformforge serve
platformforge lab start linux-navigation
platformforge lab validate linux-navigation
platformforge lab reset linux-navigation
platformforge lab stop linux-navigation
```

Check whether a lab is already running:

```bash
curl http://127.0.0.1:8080/api/labs/linux-navigation/status
```

The server binds to loopback by default. Labs drop Linux capabilities, enable `no-new-privileges`, enforce CPU/memory/PID limits, restrict temporary storage, and never mount the host Docker socket. Review third-party course packs before running them.

The original `platform-foundations` course includes five guided labs and an incident capstone spanning Linux, Git, containers, networking, Kubernetes, and incident response.

Run `make test`, `make lint`, `make smoke`, and `make e2e` (Playwright; install browser deps first). See [development setup](docs/development.md) and [lab authoring](docs/authoring.md).

Platform code is Apache-2.0. Adapted 90DaysOfDevOps content belongs in `content/90days` under CC BY-NC-SA 4.0 with attribution.
