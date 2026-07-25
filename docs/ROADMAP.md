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
| Foundations | Linux, terminal, 90Days bridges, networking tickets, Git | 23 labs |
| Containers | Docker | 3 labs |
| App development | Go, Python, SQL | 11 labs |
| Kubernetes | Core + CKA drills | 12 labs (live k3d) |
| Delivery | CI/CD, observability, leadership, portfolio | 11 labs + 4 capstones |
| Platform Engineering | Data, compliance, bare metal, AWS sims | 10 labs |

### Lab inventory (69 interactive + 4 capstones)

**Linux** — `linux-shell-basics`, `linux-navigation`, `linux-pipelines`, `linux-filesystems`

**Terminal** — `terminal-tmux`, `terminal-neovim`

**90Days bridges** — `bash-automation-basics`, `docker-tagging-basics`, `kubernetes-yaml-basics`, `networking-basics`, `monitoring-basics`, `cloud-tagging-basics`, `secrets-basics`

**Datacenter networking (PatchLab-inspired)** — `net-vlan-access`, `net-default-gateway`, `net-firewall-acl`, `net-static-nat`, `net-traceroute-path`, `net-sandbox`

**Git** — `git-recovery`, `git-branching`, `git-rebase-reset`, `git-remotes`

**Docker** — `docker-debugging`, `docker-networking`, `docker-volumes`

**HTTP & Go** — `http-health-server`, `http-json-api`, `go-testing-basics`

**Python & SQL** — `python-health-server`, `python-json-api`, `python-async-service`, `python-packaging-basics`, `sql-query-basics`, `sql-joins-basics`, `sql-indexes-explain`, `sql-window-functions`

**Kubernetes** — `kubernetes-deploy`, `kubernetes-service`, `kubernetes-configmap`, `kubernetes-scaling`, `kubernetes-namespaces`

**CKA-style** — `kubernetes-networkpolicy`, `kubernetes-rbac`, `kubernetes-troubleshooting`, `kubernetes-ingress`, `kubernetes-storage`, `kubernetes-backup-restore`, `kubernetes-etcd-snapshot`

**CI/CD & Observability** — `cicd-pipeline-fix`, `cicd-security-scan`, `observability-structured-logs`, `observability-metrics-alerts`

**Leadership** — `leadership-adr`, `leadership-postmortem`

**Portfolio project** — `portfolio-payments-api`, `portfolio-containerize`, `portfolio-ship-k8s`, `portfolio-showcase`

**Data platform** — `postgresql-pitr`, `kafka-debezium-cdc`

**Compliance** — `compliance-masking`, `dora-evidence`, `pci-dss-basics`

**Bare metal** — `proxmox-planning`, `debian-platform-baseline`

**AWS local sims** — `aws-iam-basics`, `aws-s3-basics`, `aws-vpc-basics`

**Capstones** — `incident-capstone`, `payments-reliability-capstone`, `compliance-release-capstone`, `platform-landing-zone-capstone`

## Boot.dev modules → PlatformForge status

| Boot.dev course | Status |
|-----------------|--------|
| Learn Linux | ✅ 4 labs + terminal mastery |
| Learn Git | ✅ 4 labs |
| Learn Docker | ✅ 3 labs |
| Learn Kubernetes | ✅ 5 labs + 7 CKA drills |
| Learn CI/CD | ✅ 2 labs |
| Learn Logging & Observability | ✅ 2 labs |
| Learn AWS | ✅ 3 local simulation labs |
| Learn HTTP Servers | ✅ Go + Python health/JSON labs |
| Learn Go | ✅ `go-testing-basics` (+ HTTP labs in Go) |
| Learn Python / SQL | ✅ packaging + joins/indexes/window + async service |
| Portfolio projects | ✅ 4-lab payments demo path |

## Your 5-year roadmap → PlatformForge phases

| Year / Phase | Topic | PlatformForge plan |
|--------------|-------|-------------------|
| 2026 H2 | Proxmox/Debian platform execution | ✅ `proxmox-planning`, `debian-platform-baseline` |
| 2026 H2 | Terminal mastery (Neovim/Tmux) | ✅ `terminal-tmux`, `terminal-neovim` |
| 2026–2027 | PostgreSQL PITR | ✅ `postgresql-pitr` |
| 2027 | Kafka CDC / Debezium | ✅ `kafka-debezium-cdc` |
| 2027 | Compliance masking | ✅ `compliance-masking` |
| 2027–2028 | CKA prep | ✅ through etcd snapshot runbooks |
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
| Networking | `networking-basics` + PatchLab-inspired ticket ladder |
| Monitoring | `monitoring-basics` |
| Cloud | `cloud-tagging-basics` + AWS local sims |
| Security | `secrets-basics` |
| CI/CD culture | `cicd-pipeline-fix`, `cicd-security-scan` |

## Ideas borrowed from PatchLab

[PatchLab](https://github.com/cozyGarage/patchlab) is the companion CCNA/datacenter rack trainer. PlatformForge reuses its pedagogy, not the visual rack engine:

1. **Ticket-first briefs** — each networking lab ships a `TICKET.md` with symptom + constraint
2. **Tip codes** — lessons name failure modes (`VLAN_MISMATCH`, `ACL_ORDER`, `NO_ROUTE`)
3. **Short progressive missions** — one network idea per lab, chained prerequisites
4. **Action → truth → tip** — validators assert end state; hints decode why it failed

UX now shipped from PatchLab:

1. **Ghost hints** — after every 2 failed validates on a task, the next authored hint auto-reveals
2. **Debrief/stars** — correctness / speed / cleanliness scored on pass; dashboard shows totals
3. **Sandbox unlock gates** — `net-sandbox` + path module unlock after 3 networking tickets; prerequisites enforced on start

## Next authoring priorities

1. Learning Path UX: estimated remaining time + clearer tip-code chips in validation
2. Portfolio polish (real image build/push when registry tooling is available)
3. Live etcdctl against control-plane when privileged lab runtime allows
4. Terraform / IaC local planning labs

## How to add a lab

1. Create `content/<pack>/<lab-id>/lab.yaml` + `lesson.md`
2. Register the lab ID in `content/paths/devops-engineer.yaml`
3. Run `make lint` and `make test`
