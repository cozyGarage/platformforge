# Plan a Chaos Experiment

Chaos experiments prove a steady-state hypothesis under a controlled failure — not random outages.

## Why it matters

FinTech platforms need evidence that a single replica loss does not burn the month’s error budget before a real incident does.

## Ticket focus

1. `payments-api-pod-kill` experiment sketch
2. Hypothesis / blast radius / abort write-up

## Tip codes

- `ONE_FAILURE` — start with a single-pod kill
- `STEADY_STATE` — declare healthy before inject
- `ABORT_BUDGET` — abort on burn/freeze/rollback triggers
