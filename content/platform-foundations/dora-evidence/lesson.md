# Produce DORA evidence from deploy logs

DORA metrics (deployment frequency, lead time, change fail rate, MTTR) are how platform orgs prove delivery health to leadership and auditors.

## Data

- `logs/deploys.log` — 4 production deploys
- `logs/incidents.log` — 1 rollback tied to `1.1.1`

## Tasks

1. Write `/workspace/evidence/dora.env` with:
   - `deployment_frequency=4`
   - `change_fail_rate=25` (1 failed / 4 deploys × 100)
2. Write `/workspace/evidence/summary.md` with headings **Deployment frequency** and **Change fail rate**
