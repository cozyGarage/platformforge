# Plan FinOps Cost Guardrails

FinOps guardrails make spend attributable and actionable: labels, budgets, and rightsizing against real usage.

## Why it matters

Unlabeled workloads hide who pays. Oversized requests burn cluster capacity before the invoice arrives.

## Ticket focus

1. Required cost labels + enforcement
2. Monthly budget alert (~80%)
3. Rightsizing note from usage.env

## Tip codes

- `COST_LABELS` — team + service + env
- `LABEL_ENFORCE` — admission / scorecard / CI
- `BUDGET_ALERT` — warn before the month is gone
- `RIGHTSIZE` — requests should track usage
