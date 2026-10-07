# Plan Logs-Metrics-Traces Correlation

An alert without a jump path is just noise. Correlation is how on-call moves alert → trace → logs in one story.

## Why it matters

Separate tools without shared IDs force copy-paste detective work during a payments outage.

## Ticket focus

1. Identity contract: trace_id, service, exemplars, anti-patterns
2. Alert → Trace → Logs lookup runbook

## Tip codes

- `CORRELATE_IDS` — shared trace_id; metrics use exemplars
- `ALERT_TO_TRACE` — page → labels/exemplar → filtered logs
