# Author a Helm Chart Skeleton

Helm packages Kubernetes manifests as a versioned chart with values-driven templates.

## Why it matters

Raw YAML copies drift. A chart gives you one release artifact, reproducible upgrades, and a clear place for env-specific values.

## Ticket focus

1. `Chart.yaml` metadata (`apiVersion: v2`)
2. `values.yaml` defaults (`replicaCount`, `image`)
3. `templates/deployment.yaml` reading `.Values`

## Tip codes

- `CHART_API` — Helm 3 charts use `apiVersion: v2`
- `VALUES_IMAGE` — keep repository and tag under `image:`
- `TPL_VALUES` — templates reference `.Values.*`
