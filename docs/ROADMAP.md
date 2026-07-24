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

| Phase | Focus | PlatformForge labs |
|-------|-------|-------------------|
| Foundations | Linux, terminal, 90Days bridges, Git | 13 labs |
| Containers | Docker | 3 labs |
| HTTP & Go | Health, JSON API, tests | 3 labs |
| Kubernetes | Core + CKA drills | 11 labs (live k3d) |
| Delivery | CI/CD, observability, leadership | 7 labs + incident capstone |
| Platform Engineering | Data, compliance, bare metal, AWS sims | 10 labs |

### Lab inventory (46 interactive + 1 capstone)

**Linux** — `linux-shell-basics`, `linux-navigation`, `linux-pipelines`, `linux-filesystems`

**Terminal** — `terminal-tmux`, `terminal-neovim`

**90Days bridges** — `bash-automation-basics`, `docker-tagging-basics`, `kubernetes-yaml-basics`

**Git** — `git-recovery`, `git-branching`, `git-rebase-reset`, `git-remotes`

**Docker** — `docker-debugging`, `docker-networking`, `docker-volumes`

**HTTP & Go** — `http-health-server`, `http-json-api`, `go-testing-basics`

**Kubernetes** — `kubernetes-deploy`, `kubernetes-service`, `kubernetes-configmap`, `kubernetes-scaling`, `kubernetes-namespaces`

**CKA-style** — `kubernetes-networkpolicy`, `kubernetes-rbac`, `kubernetes-troubleshooting`, `kubernetes-ingress`, `kubernetes-storage`, `kubernetes-backup-restore`

**CI/CD & Observability** — `cicd-pipeline-fix`, `cicd-security-scan`, `observability-structured-logs`, `observability-metrics-alerts`

**Leadership** — `leadership-adr`, `leadership-postmortem`

**Data platform** — `postgresql-pitr`, `kafka-debezium-cdc`

**Compliance** — `compliance-masking`, `dora-evidence`, `pci-dss-basics`

**Bare metal** — `proxmox-planning`, `debian-platform-baseline`

**AWS local sims** — `aws-iam-basics`, `aws-s3-basics`, `aws-vpc-basics`

**Capstone** — `incident-capstone`

## Boot.dev modules → PlatformForge status

| Boot.dev course | Status |
|-----------------|--------|
| Learn Linux | ✅ 4 labs + terminal mastery |
| Learn Git | ✅ 4 labs |
| Learn Docker | ✅ 3 labs |
| Learn Kubernetes | ✅ 5 labs + 6 CKA drills |
| Learn CI/CD | ✅ 2 labs |
| Learn Logging & Observability | ✅ 2 labs |
| Learn AWS | ✅ 3 local simulation labs |
| Learn HTTP Servers | ✅ `http-health-server`, `http-json-api` |
| Learn Go | ✅ `go-testing-basics` (+ HTTP labs in Go) |
| Learn Python / SQL | 📋 Planned |
| Portfolio projects | 📋 Next priority |

## Your 5-year roadmap → PlatformForge phases

| Year / Phase | Topic | PlatformForge plan |
|--------------|-------|-------------------|
| 2026 H2 | Proxmox/Debian platform execution | ✅ `proxmox-planning`, `debian-platform-baseline` |
| 2026 H2 | Terminal mastery (Neovim/Tmux) | ✅ `terminal-tmux`, `terminal-neovim` |
| 2026–2027 | PostgreSQL PITR | ✅ `postgresql-pitr` |
| 2027 | Kafka CDC / Debezium | ✅ `kafka-debezium-cdc` |
| 2027 | Compliance masking | ✅ `compliance-masking` |
| 2027–2028 | CKA prep | ✅ NetworkPolicy through backup/restore |
| 2028+ | DORA / PCI-DSS evidence | ✅ `dora-evidence`, `pci-dss-basics` |
| 2028+ | Leadership capstones | ✅ `leadership-adr`, `leadership-postmortem` |

## 90DaysOfDevOps attribution

Adapted content lives under [`content/90days/`](../content/90days/) (CC BY-NC-SA 4.0).

| 90Days topic area | PlatformForge lab |
|-------------------|-------------------|
| Linux fundamentals | `linux-shell-basics`, `bash-automation-basics` |
| Git / version control | `git-recovery`, `git-branching` |
| Containers | `docker-tagging-basics` + docker module |
| Kubernetes | `kubernetes-yaml-basics` + kubernetes/CKA modules |
| CI/CD culture | `cicd-pipeline-fix`, `cicd-security-scan` |

## Next authoring priorities

1. Portfolio project labs that stitch HTTP + Docker + Kubernetes into a public demo path
2. More 90Days day conversions (networking, monitoring, cloud days)
3. Learn Python / SQL service labs
4. Optional etcd snapshot lab when local control-plane tooling allows

## How to add a lab

1. Create `content/<pack>/<lab-id>/lab.yaml` + `lesson.md`
2. Register the lab ID in `content/paths/devops-engineer.yaml`
3. Run `make lint` and `make test`
