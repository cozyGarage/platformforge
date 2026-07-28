# Define an Availability SLO

An SLO turns a measurable SLI into a target over a window — the contract on-call and product share.

## Why it matters

Without a written target and window, "is payments-api healthy?" becomes an opinion instead of budget math.

## Ticket focus

1. Compute SLI from success/total
2. Target `99.9%`
3. Window `30` days

## Tip codes

- `SLI_RATIO` — successful requests / total requests
- `SLO_TARGET` — explicit target like 99.9% availability
- `WINDOW_30D` — measurement window so budgets are comparable
