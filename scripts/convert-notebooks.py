#!/usr/bin/env python3
"""Regenerate the notebook-derived readings listed in UNITS (provenance + reproducible conversion).

    python3 scripts/convert-notebooks.py [NOTEBOOK_REPO]   # default: ../courses-notebook

Each unit is converted with nb2reading.py into content/theory/<id>/. Review the output: conversion
is a starting point. Re-running overwrites lesson.md/reading.yaml, so hand edits to a converted
unit belong in git history, not in this list (or drop the unit from UNITS once it is hand-owned).
"""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
NOTEBOOKS = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT.parent / "courses-notebook"

# id, title, course, notebook, mcq question numbers, summary
UNITS = [
    # SRE and observability (SE303)
    ("sre-fundamentals", "SRE Fundamentals: SLIs, SLOs and Error Budgets", "SE303_observability_sre", "01_sre_fundamentals", [1, 5, 6, 17, 18],
     "How an SLI becomes an SLO, what an error budget is, and why exhausting it changes how a team ships."),
    ("metrics-and-monitoring", "Metrics and Monitoring", "SE303_observability_sre", "02_metrics_and_monitoring", [2, 3, 4, 12, 15],
     "The four golden signals, Prometheus metric types, and the PromQL patterns for rates and latency percentiles."),
    ("structured-logging", "Structured Logging", "SE303_observability_sre", "03_structured_logging", [9, 19],
     "Why production logs should be structured, how severities and correlation ids work, and what to log."),
    ("distributed-tracing", "Distributed Tracing", "SE303_observability_sre", "04_distributed_tracing", [7, 13, 16],
     "Following one request across services with traces and spans, and where tracing beats metrics and logs."),
    ("alerting-and-oncall", "Alerting and On-Call", "SE303_observability_sre", "05_alerting_and_oncall", [11, 14, 20],
     "Alert fatigue, symptom versus cause alerts, burn-rate alerting, runbooks and incident severity."),
    ("incident-management", "Incident Management and Blameless Postmortems", "SE303_observability_sre", "06_incident_management", [8],
     "Roles and timeline of an incident, and how a blameless postmortem turns it into durable fixes."),
    ("chaos-engineering", "Chaos Engineering", "SE303_observability_sre", "07_chaos_engineering", [10],
     "Steady-state hypotheses, blast-radius control and game days: breaking things on purpose, safely."),
    ("slo-alerting-case-study", "Case Study: SLO Alerting with Error-Budget Burn Rate", "SE303_observability_sre", "08_case_study", [],
     "Design and test a multi-window burn-rate alerting system for an SLO from first principles."),
    # Git (ENV202)
    ("git-object-model", "The Git Object Model", "ENV202_git_mastery", "01_git_object_model", [1, 2, 8, 12],
     "Blobs, trees, commits and tags; what a branch really is; and how the index relates to the working tree."),
    ("git-branching-strategies", "Branching Strategies", "ENV202_git_mastery", "02_branching_strategies", [4, 6, 16],
     "GitHub Flow, Git Flow and trunk-based development, plus merge, fast-forward and fetch versus pull."),
    ("git-rebasing-and-history", "Rebasing and Rewriting History Safely", "ENV202_git_mastery", "03_rebasing_and_history", [3, 5, 7, 9, 14, 18],
     "Rebase, squash, amend, reset and revert, the golden rule of rebasing, and recovering lost commits."),
    ("git-hooks-and-automation", "Git Hooks and Automation", "ENV202_git_mastery", "04_git_hooks", [13],
     "Client and server hooks for linting, commit checks and automation, and their limits."),
    ("github-workflow", "GitHub Workflow and Pull Requests", "ENV202_git_mastery", "05_github_workflow", [],
     "Fork, branch, review and merge: the pull-request workflow and branch protection."),
    ("monorepos-and-submodules", "Monorepos and Submodules", "ENV202_git_mastery", "06_monorepos_and_submodules", [],
     "Trade-offs between monorepos, polyrepos and submodules, and the tooling each needs."),
    # SQL (DS301)
    ("sql-joins-and-subqueries", "Joins and Subqueries", "DS301_databases_sql", "03_joins_subqueries", [5],
     "Inner, outer and cross joins, correlated subqueries and when each is the right tool."),
    ("window-functions-in-sql", "Window Functions", "DS301_databases_sql", "09_window_functions", [6, 7, 19],
     "Ranking, running totals and lag/lead without collapsing rows: PARTITION BY, ORDER BY and frames."),
    ("sql-cohort-analysis", "Cohort Analysis: Retention, Churn and LTV", "DS301_databases_sql", "10_cohort_analysis", [],
     "Group users by first activity and measure retention and lifetime value over time in plain SQL."),
    ("sql-funnel-analysis", "Funnel Analysis", "DS301_databases_sql", "11_funnel_analysis", [],
     "Measure step-to-step conversion and find the drop-off in an ordered event funnel."),
    ("sql-time-series", "Time-Based Analysis in SQL", "DS301_databases_sql", "12_time_series_sql", [],
     "Date bucketing, moving averages and period-over-period comparisons."),
    ("sql-query-tuning-case-study", "Case Study: Tuning a Slow Analytical Query", "DS301_databases_sql", "08_case_study", [10, 12, 16, 20],
     "Read an execution plan, find why a query is slow, and fix it with indexes and rewrites."),
    # Kubernetes (SYS401)
    ("k8s-autoscaling-hpa-vpa", "Autoscaling: HPA and VPA", "SYS401_kubernetes_devops", "03_hpa_vpa_scaling", [6, 7, 18, 20],
     "The HPA control loop, stabilisation windows, multi-metric scaling and what the VPA recommends."),
    ("k8s-namespaces-and-rbac", "Namespaces, RBAC and Resource Quotas", "SYS401_kubernetes_devops", "04_namespaces_rbac", [8, 16],
     "Isolation with namespaces, least-privilege Roles and bindings, ResourceQuota and LimitRange."),
    ("gitops-with-argocd", "GitOps with Argo CD", "SYS401_kubernetes_devops", "06_gitops_argocd", [14, 15],
     "Desired state in Git, the reconcile loop, sync policies, prune and ApplicationSet promotion."),
    ("k8s-split-brain-case-study", "Case Study: Split-Brain Recovery in a Leader-Elected Controller", "SYS401_kubernetes_devops", "08_case_study", [],
     "Leader election, fencing tokens and idempotency: recovering from two controllers acting at once."),
    # FinOps and multi-cloud (CLD303, CLD302)
    ("finops-foundations", "FinOps Foundations", "CLD303_finops_cloud_cost", "01_finops_foundations", [1, 2],
     "The inform, optimise, operate lifecycle and the cultural rules that make cloud cost an engineering concern."),
    ("finops-compute-optimization", "Compute Cost Optimization", "CLD303_finops_cloud_cost", "03_compute_optimization", [3, 4, 5, 12, 13],
     "Rightsizing, spot, savings plans and commitment coverage, in the order that avoids waste."),
    ("finops-waste-case-study", "Case Study: A SaaS Company Discovers 40% Cloud Waste", "CLD303_finops_cloud_cost", "08_case_study", [],
     "Find, quantify and remove waste in a real-shaped bill, and report the savings."),
    ("multicloud-identity", "Identity and Access Across Clouds", "CLD302_multi_cloud_architecture", "03_identity_and_access", [7, 8, 13],
     "Centralised identity, workload identity federation and how IAM evaluation differs between providers."),
    ("multicloud-migration-patterns", "Migration Patterns", "CLD302_multi_cloud_architecture", "07_migration_patterns", [9, 10],
     "Rehost, replatform and refactor, and why the strangler fig beats a big-bang migration."),
    ("multicloud-lock-in-case-study", "Case Study: A SaaS Leaves Single-Cloud Lock-In", "CLD302_multi_cloud_architecture", "08_case_study", [],
     "Plan identity federation, data portability and a staged cutover away from a single provider."),
]


def main() -> None:
    for uid, title, course, notebook, questions, summary in UNITS:
        base = NOTEBOOKS / "courses" / course
        cmd = [
            sys.executable, str(ROOT / "scripts" / "nb2reading.py"), str(base / f"{notebook}.ipynb"),
            "--id", uid, "--title", title, "--summary", summary,
            "--source", f"courses-notebook {course} {notebook}",
            "--out", str(ROOT / "content" / "theory" / uid),
        ]
        if questions:
            cmd += ["--mcq", str(base / "mcq.ipynb"), "--questions", ",".join(map(str, questions))]
        subprocess.run(cmd, check=True, stdout=subprocess.DEVNULL)
    print(f"converted {len(UNITS)} units")


if __name__ == "__main__":
    main()
