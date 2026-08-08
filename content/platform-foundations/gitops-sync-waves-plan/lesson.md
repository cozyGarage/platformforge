# Plan GitOps Sync Waves and Health Gates

Fleet apply without waves is a simultaneous outage machine. Order the sync, gate on health, fail closed.

## Why it matters

us-east canary success should be required before eu-west takes the same digest. Otherwise every region fails together.

## Ticket focus

1. Six-line wave order: prereqs → canary → gate → follow → fail-closed
2. Health / Gate / Rollback notes

## Tip codes

- `SYNC_WAVES` — prereqs, canary region, then follow regions
- `HEALTH_GATE` — advance only after ready/smoke/SLO checks
- `WAVE_ROLLBACK` — failed canary reverts digest; do not continue the fleet
