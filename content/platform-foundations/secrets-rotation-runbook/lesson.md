# Write a Secrets Rotation Runbook

When a credential may be leaked, rotate the backend first, refresh sync, then verify every consumer.

## Why it matters

Deleting a Kubernetes Secret alone does nothing if the vault still serves the old value — or if pods still mount the old material.

## Ticket focus

1. Detect → Rotate → Refresh → Verify
2. Dual-write / overlap window
3. Three-point verify checklist

## Tip codes

- `ROTATE_BACKEND` — rotate in vault/backend first
- `REFRESH_ESO` — refresh ExternalSecret after backend rotation
- `OVERLAP_WINDOW` — dual-write so old and new both work briefly
- `VERIFY_THREE` — store sync, secret material, app health
