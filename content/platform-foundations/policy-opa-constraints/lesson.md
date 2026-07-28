# Plan OPA Gatekeeper Constraints

Gatekeeper splits policy into a reusable ConstraintTemplate (Rego) and concrete Constraints that match cluster objects.

## Why it matters

Finance and platform teams need the same rule applied consistently — Audit first to measure blast radius, then Deny.

## Ticket focus

1. ConstraintTemplate for `cost-center`
2. Constraint targeting Pods/Deployments
3. Audit vs Deny mode notes

## Tip codes

- `OPA_TEMPLATE` — template owns Rego + CRD shape
- `OPA_CONSTRAINT` — constraint instantiates the template
- `AUDIT_FIRST` — Audit before Deny in prod rollouts
