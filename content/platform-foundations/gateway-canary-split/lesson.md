# Plan Gateway Canary Traffic Splits

Weighted `backendRefs` let you shift a small percentage of traffic to a canary Service before promoting.

## Why it matters

All-or-nothing rollouts hide regressions until everyone feels them. A 10% canary with clear rollback weights keeps blast radius small.

## Ticket focus

1. HTTPRoute 90/10 split
2. Metrics → Promote → Rollback checklist

## Tip codes

- `WEIGHT_SPLIT` — weights should sum to 100
- `CANARY_NAME` — distinct canary Service name
- `CANARY_METRICS` — watch error rate/latency first
- `CANARY_ROLLBACK` — weight 0 / detach canary
