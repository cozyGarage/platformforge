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
| IaC | Terraform + Ansible planning | 6 labs |
| App development | Go, Python, SQL | 11 labs |
| Kubernetes | Core + CKA + policy + Gateway API | 16 labs |
| Delivery | CI/CD, GitOps, Helm, secrets ops, observability, SRE, leadership, portfolio | 22 labs + 7 capstones |
| Platform Engineering | Platform product, data, compliance, bare metal, AWS sims | 16 labs |

### Lab inventory (97 interactive + 7 capstones)

**Linux** — `linux-shell-basics`, `linux-navigation`, `linux-pipelines`, `linux-filesystems`

**Terminal** — `terminal-tmux`, `terminal-neovim`

**90Days bridges** — `bash-automation-basics`, `docker-tagging-basics`, `kubernetes-yaml-basics`, `networking-basics`, `monitoring-basics`, `cloud-tagging-basics`, `secrets-basics`

**Datacenter networking (PatchLab-inspired)** — `net-vlan-access`, `net-default-gateway`, `net-firewall-acl`, `net-static-nat`, `net-traceroute-path`, `net-sandbox`

**Git** — `git-recovery`, `git-branching`, `git-rebase-reset`, `git-remotes`

**Docker** — `docker-debugging`, `docker-networking`, `docker-volumes`

**Terraform / IaC** — `terraform-basics`, `terraform-modules`, `terraform-state-backend`

**Ansible** — `ansible-inventory-basics`, `ansible-playbook-basics`, `ansible-handlers-idempotency`

**HTTP & Go** — `http-health-server`, `http-json-api`, `go-testing-basics`

**Python & SQL** — `python-health-server`, `python-json-api`, `python-async-service`, `python-packaging-basics`, `sql-query-basics`, `sql-joins-basics`, `sql-indexes-explain`, `sql-window-functions`

**Kubernetes** — `kubernetes-deploy`, `kubernetes-service`, `kubernetes-configmap`, `kubernetes-scaling`, `kubernetes-namespaces`

**CKA-style** — `kubernetes-networkpolicy`, `kubernetes-rbac`, `kubernetes-troubleshooting`, `kubernetes-ingress`, `kubernetes-storage`, `kubernetes-backup-restore`, `kubernetes-etcd-snapshot`

**Policy as code** — `policy-kyverno-basics`, `policy-opa-constraints`

**Gateway API** — `gateway-api-http-route`, `gateway-canary-split`

**CI/CD & Observability** — `cicd-pipeline-fix`, `cicd-security-scan`, `observability-structured-logs`, `observability-metrics-alerts`

**Reliability / SRE** — `slo-definition-basics`, `error-budget-policy`, `burn-rate-alerts`, `chaos-experiment-basics`, `gameday-budget-freeze`, `oncall-handoff-basics`

**GitOps** — `gitops-manifest-sync`, `gitops-kustomize-overlay`

**Helm** — `helm-chart-basics`, `helm-values-overrides`

**Secrets ops** — `secrets-external-operator`, `secrets-rotation-runbook`

**Leadership** — `leadership-adr`, `leadership-postmortem`

**Portfolio project** — `portfolio-payments-api`, `portfolio-containerize`, `portfolio-ship-k8s`, `portfolio-showcase`

**Data platform** — `postgresql-pitr`, `kafka-debezium-cdc`, `redis-ha-failover-plan`

**Compliance** — `compliance-masking`, `dora-evidence`, `pci-dss-basics`, `soc2-change-evidence`

**Bare metal** — `proxmox-planning`, `debian-platform-baseline`, `rack-capacity-planning`

**AWS local sims** — `aws-iam-basics`, `aws-s3-basics`, `aws-vpc-basics`

**Platform product** — `idp-golden-path-basics`, `finops-cost-guardrails`, `multi-env-promotion-adr`

**Capstones** — `incident-capstone`, `payments-reliability-capstone`, `compliance-release-capstone`, `platform-landing-zone-capstone`, `gitops-delivery-capstone`, `reliability-gameday-capstone`, `platform-product-capstone`

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
4. **Path remaining-time + tip-code chips** — phase/path ETA and failure tip chips on validate
5. **Next-lab CTA + tip glossary** — after pass, continue along the path; open a lab tip glossary drawer
6. **Tip chips → glossary** — failed-validate tip chips jump-open the matching glossary entry
7. **Continue where you left off** — Learning Path and Catalog heroes link to the next incomplete unlocked lab

## Roadmap phases

### Phase A — Close the reliability loop (shipped)
1. `oncall-handoff-basics` — severity ladder, open pages, escalation roles
2. `reliability-gameday-capstone` — burn → freeze → chaos → postmortem evidence pack
3. Catalog Continue CTA (shared with Learning Path)

### Phase B — Platform product skills (shipped)
1. `idp-golden-path-basics` — service template, scorecard, self-service
2. `finops-cost-guardrails` — cost labels, budget alert, rightsizing
3. `multi-env-promotion-adr` — GitOps/Helm digest promotion ADR + checklist

### Phase C — Runtime unlock (blocked until tooling; shown as comingSoon in path UI)
1. Portfolio real image build/push (registry available)
2. Live `etcdctl` against control-plane (privileged runtime)
3. Live Kyverno / Gateway apply on k3d (CRDs in lab image)

### Phase D — Curriculum polish (shipped this cycle)
1. Platform Engineering depth — `redis-ha-failover-plan`, `soc2-change-evidence`, `rack-capacity-planning`
2. Path `comingSoon` slots for Phase C on portfolio / etcd / policy / gateway modules
3. Site/Pages rebuilds from path on main CI deploy
4. `platform-product-capstone` — stitches IDP + FinOps + promotion ADR

### Phase E — Next horizons
1. Unlock Phase C when registry / privileged etcd / CRD-enabled k3d images exist
2. Observability depth (tracing / SLI recording rules planning)
3. Multi-cluster / fleet GitOps planning drills
## How to add a lab

1. Create `content/<pack>/<lab-id>/lab.yaml` + `lesson.md`
2. Register the lab ID in `content/paths/devops-engineer.yaml`
3. Run `make lint` and `make test`
