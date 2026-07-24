# PlatformForge Curriculum Roadmap

This document maps **Boot.dev DevOps**, **90DaysOfDevOps**, and your **5-year FinTech platform roadmap** into PlatformForge modules.

## Design principles (Boot.dev-inspired)

1. **Short chapters** — each lab is 20–40 minutes with one clear objective
2. **Hands-on first** — read briefly, fix something real, validate automatically
3. **Progressive path** — prerequisites enforce order; the UI shows a structured path
4. **Projects between courses** — capstones tie multiple skills together
5. **Coming soon slots** — roadmap modules exist before labs are authored

## DevOps Engineer Path (implemented)

See [`content/paths/devops-engineer.yaml`](../content/paths/devops-engineer.yaml) and the **Learning Path** page in the UI.

| Phase | Boot.dev equivalent | PlatformForge labs |
|-------|---------------------|-------------------|
| Foundations | Learn Linux, Learn Git | 8 labs (linux + git) |
| Containers | Learn Docker | 3 labs |
| Kubernetes | Learn Kubernetes | 5 labs (live k3d) |
| Delivery | CI/CD + Observability | 5 labs + capstone |
| Platform Engineering | Data + roadmap | 1 lab (`postgresql-pitr`) + coming soon |

### Lab inventory (22 interactive + 1 capstone)

**Linux** — `linux-shell-basics`, `linux-navigation`, `linux-pipelines`, `linux-filesystems`

**Git** — `git-recovery`, `git-branching`, `git-rebase-reset`, `git-remotes`

**Docker** — `docker-debugging`, `docker-networking`, `docker-volumes`

**Kubernetes (k3d)** — `kubernetes-deploy`, `kubernetes-service`, `kubernetes-configmap`, `kubernetes-scaling`, `kubernetes-namespaces`

**CI/CD & Observability** — `cicd-pipeline-fix`, `cicd-security-scan`, `observability-structured-logs`, `observability-metrics-alerts`

**Data platform** — `postgresql-pitr`

**Capstone** — `incident-capstone`

## Boot.dev modules → PlatformForge status

| Boot.dev course | Status |
|-----------------|--------|
| Learn Linux | ✅ 4 labs |
| Learn Git | ✅ 4 labs |
| Learn Docker | ✅ 3 labs |
| Learn Kubernetes | ✅ 5 labs |
| Learn CI/CD | ✅ 2 labs |
| Learn Logging & Observability | ✅ 2 labs |
| Learn Go / Python / SQL | 📋 Planned (`postgresql-pitr` covers recovery drills) |
| Learn AWS | 📋 Planned (local simulations first) |
| Learn HTTP Servers | 📋 Planned |
| Portfolio projects | 📋 After core path |

## Your 5-year roadmap → PlatformForge phases

From your FinTech architect roadmap (`gemini-code` context):

| Year / Phase | Topic | PlatformForge plan |
|--------------|-------|-------------------|
| 2026 H2 | Proxmox/Debian platform execution | `proxmox-planning` (config exercises, homelab provider later) |
| 2026 H2 | Terminal mastery (Neovim/Tmux) | Optional editor labs after Linux path |
| 2026–2027 | PostgreSQL PITR | ✅ `postgresql-pitr` (WAL replay simulation) |
| 2027 | Kafka CDC / Debezium | `kafka-debezium-cdc` |
| 2027 | Compliance masking | `compliance-masking` |
| 2027–2028 | CKA prep | Expand Kubernetes track (CKA-style scenarios) |
| 2028+ | DORA / PCI-DSS evidence | `dora-evidence`, `pci-dss-basics` |
| 2028+ | Leadership capstones | Architecture decision records, postmortems |

## 90DaysOfDevOps attribution

Adapted content lives under [`content/90days/`](../content/90days/) (CC BY-NC-SA 4.0).

| 90Days topic area | PlatformForge lab |
|-------------------|-------------------|
| Linux fundamentals | `linux-shell-basics` |
| Git / version control | `git-recovery`, `git-branching` |
| Containers | docker module |
| Kubernetes | kubernetes module |
| CI/CD culture | `cicd-pipeline-fix`, `cicd-security-scan` |

## Next authoring priorities

1. `kafka-debezium-cdc` — change-data-capture basics
2. `compliance-masking` — PII redaction drills
3. `proxmox-planning` — bare-metal platform checklist exercises
4. More 90Days day-by-day conversions (scenario rewrite + validators)

## How to add a lab

1. Create `content/<pack>/<lab-id>/lab.yaml` + `lesson.md`
2. Register the lab ID in `content/paths/devops-engineer.yaml`
3. Run `make lint` and `make test`
