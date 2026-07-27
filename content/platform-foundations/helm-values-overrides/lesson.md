# Plan Helm Values Overrides

Charts stay reusable when environment deltas live in layered values files — not forked charts.

## Why it matters

Copy-pasted chart trees per env drift. `-f values-prod.yaml` keeps the chart shared and the diff reviewable.

## Ticket focus

1. `values-dev.yaml` / `values-prod.yaml`
2. Prod digests over floating tags
3. Prefer `-f` over `--set`; never put secrets in values

## Tip codes

- `ENV_VALUES` — env deltas in `values-<env>.yaml`
- `PROD_DIGEST` — prod pins digests, not `latest`
- `PREFER_F` — prefer `-f` over long `--set` chains
- `SET_EPHEMERAL` — `--set` for break-glass / CI-only
