# Plan a GitOps Manifest Sync Loop

GitOps means the cluster continuously converges to what Git declares — not the other way around.

## Why it matters

Imperative `kubectl apply` from laptops creates snowflakes. Declarative sync makes audits, rollbacks, and multi-env promotion explainable.

## Ticket focus

1. What belongs in the desired-state repo
2. How drift is detected and reconciled
3. The watch → compare → apply → report loop

## Tip codes

- `GITOPS_SOURCE` — Git holds desired state; clusters are projections
- `DRIFT_RECONCILE` — drift signals a sync toward Git
- `NO_SNOWFLAKE` — untracked live-only edits are not the primary fix
- `SYNC_LOOP` — watch → compare → apply → report
- `SOURCE_OF_TRUTH` — one declared source; many projected environments
