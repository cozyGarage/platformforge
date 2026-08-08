# Plan Multi-Cluster Fleet GitOps

One cluster GitOps is a repo. Multi-cluster GitOps is a fleet — layout, selectors, and wave order matter.

## Why it matters

Payments in us-east and eu-west without selectors and waves becomes copy-paste overlays and uneven rollbacks.

## Ticket focus

1. Fleet layout: shared apps + per-cluster registry + hub/control note
2. Selectors, promotion waves, and digest rollback

## Tip codes

- `FLEET_LAYOUT` — shared apps vs per-cluster registry
- `CLUSTER_SELECT` — label/selector targeting, not one-off apps forever
- `WAVE_PROMOTE` — canary/region waves before full fleet; keep digest rollback
