# Plan Redis HA Failover

Payments caches and locks need an explicit primary/replica story and a client cutover path when the primary dies.

## Why it matters

“Redis is up” is not a plan. Promotion without client cutover still looks like an outage.

## Ticket focus

1. Topology: primary, replica, failover manager
2. Trigger → Promote → Clients runbook

## Tip codes

- `REDIS_TOPO` — primary, replica, who promotes
- `FAILOVER_TRIGGER` — objective failure signals
- `CLIENT_CUTOVER` — reconnect / DNS / endpoint
