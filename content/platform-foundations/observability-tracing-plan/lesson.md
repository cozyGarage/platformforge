# Plan Distributed Tracing Coverage

Metrics tell you something is wrong. Traces tell you which hop burned the budget.

## Why it matters

Payments paths cross API, ledger, fraud, and storage. Without a span map and propagation plan, on-call greps logs while customers wait.

## Ticket focus

1. Five-line span map across payments-api and downstreams
2. Propagation + sampling + on-call lookup notes

## Tip codes

- `SPAN_MAP` — entry span and each hop you must follow
- `PROPAGATE` — carry W3C traceparent (or equivalent) across hops
- `SAMPLE_POLICY` — head/tail sampling with an explicit ratio
