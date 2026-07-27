# Plan Kustomize Overlay Promotions

Kustomize overlays isolate environment deltas on top of a shared base so promotions stay reviewable.

## Why it matters

Copy-pasted full manifests per env drift silently. Overlay discipline makes "what changed for prod?" answerable in a PR.

## Ticket focus

1. `base/` + `overlays/{dev,stage,prod}`
2. Digest promotion upward
3. Drift and hotfix backport risks

## Tip codes

- `KUST_BASE` — shared resources in base; env deltas in overlays
- `OVERLAY_DELTA` — image, replicas, resources, hosts — not a full fork
- `PROMOTE_DIGEST` — same digest upward across envs
- `NO_LATEST` — floating latest breaks audits
- `OVERLAY_DRIFT` — fight copy-paste divergence
- `HOTFIX_BACKPORT` — prod fixes must return to base/dev
