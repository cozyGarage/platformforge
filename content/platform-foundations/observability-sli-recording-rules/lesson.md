# Plan SLI Recording Rules

Burn-rate alerts should read stable SLI series — not recompute scrape math on every evaluate.

## Why it matters

Ad-hoc PromQL in paging rules drifts, gets expensive, and disagrees with the SLO dashboard. Recording rules make availability and latency SLIs first-class inputs.

## Ticket focus

1. Availability + latency recording rules for payments-api
2. SLI.md that ties rules to burn-rate / error-budget use

## Tip codes

- `RECORD_SLI` — precompute SLIs for cheap, stable alerts
- `BURN_INPUT` — burn-rate alerts consume recording-rule series
