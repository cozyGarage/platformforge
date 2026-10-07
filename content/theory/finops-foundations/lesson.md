Understand the FinOps framework lifecycle, the FOCUS billing specification, and unit economics — then build a Python billing data aggregator that simulates a cloud cost report.

## The FinOps Framework

The FinOps Foundation defines three phases executed in a continuous loop:

| Phase | Goal | Key Activities |
|---|---|---|
| **Inform** | Visibility & allocation | Cost dashboards, tagging, showback reports |
| **Optimize** | Reduce waste & maximize value | Rightsizing, RI/SP purchasing, scheduling |
| **Operate** | Embed discipline in culture | Budgets, alerts, cost reviews, KPIs |

### FinOps Personas

| Persona | Primary Concern | Key Metric |
|---|---|---|
| Engineering | Build reliable, cost-efficient systems | Cost per deployment |
| Finance | Forecast and budget cloud spend | Variance vs budget |
| Product | Understand cost per feature/user | Cost per active user |
| Leadership | Unit economics, margin | Cloud as % of revenue |

### FinOps Maturity Model

| Stage | Characteristics |
|---|---|
| **Crawl** | Reactive — basic tagging, monthly cost reviews |
| **Walk** | Proactive — showback, anomaly alerts, RI coverage |
| **Run** | Predictive — unit economics, real-time optimization, FinOps RACI |

## The FOCUS Billing Specification

**FOCUS** (FinOps Open Cost & Usage Specification) is a vendor-neutral schema for cloud billing data, standardizing field names across AWS (CUR), GCP (BigQuery billing), and Azure (Cost Management exports).

### Core FOCUS Columns

| Column | Type | Description |
|---|---|---|
| `BillingPeriodStart` | datetime | Start of billing window |
| `BillingPeriodEnd` | datetime | End of billing window |
| `ServiceName` | string | Cloud service (EC2, S3, BigQuery) |
| `ServiceCategory` | string | Compute, Storage, Network, Database |
| `ResourceId` | string | Unique resource identifier |
| `RegionName` | string | Cloud region |
| `BilledCost` | float | Actual charged amount (USD) |
| `ListCost` | float | On-demand list price before discounts |
| `EffectiveCost` | float | Amortized cost including commitments |
| `Tags` | map | Key-value resource tags |
| `ProviderName` | string | AWS, GCP, Azure |

FOCUS enables multi-cloud cost analysis without custom ETL per provider.

## Unit Economics

Unit economics connects cloud spend to business outcomes:

| Unit Metric | Formula | Target |
|---|---|---|
| Cost per Active User (CPAU) | Total cloud cost / MAU | Decreasing as you scale |
| Cost per API Request | Total cost / Request count | < $0.001 for most APIs |
| Cost per Transaction | Total cost / Transactions | Depends on domain |
| Cloud Margin | 1 − (Cloud cost / Revenue) | Should grow over time |
| Waste Ratio | Waste cost / Total cost | Target < 10% |

Without unit economics, cost reduction efforts are disconnected from business value.

```python
# Cell 1: FOCUS-compliant billing record model and data generator
from dataclasses import dataclass, field
from datetime import date, timedelta
from typing import Dict, List, Optional
import random
import json


@dataclass
class BillingRecord:
    """A single FOCUS-compliant billing line item."""
    billing_period_start: date
    billing_period_end: date
    service_name: str
    service_category: str  # Compute, Storage, Network, Database
    resource_id: str
    region_name: str
    billed_cost: float
    list_cost: float
    effective_cost: float
    provider_name: str
    tags: Dict[str, str] = field(default_factory=dict)

    @property
    def discount_pct(self) -> float:
        """Percentage discount vs list price."""
        if self.list_cost == 0:
            return 0.0
        return round((1 - self.billed_cost / self.list_cost) * 100, 2)


def generate_billing_dataset(days: int = 30, seed: int = 42) -> List[BillingRecord]:
    """Generate a synthetic FOCUS-compliant cloud billing dataset."""
    random.seed(seed)

    services = [
        ('EC2',         'Compute',  1.0,  True),
        ('RDS',         'Database', 0.8,  True),
        ('S3',          'Storage',  0.02, False),
        ('EKS',         'Compute',  0.5,  True),
        ('CloudFront',  'Network',  0.05, False),
        ('Lambda',      'Compute',  0.1,  False),
        ('ElastiCache', 'Database', 0.3,  True),
        ('DataTransfer','Network',  0.09, False),
    ]

    teams = ['platform', 'data', 'product', 'ml', 'infra']
    envs  = ['prod', 'staging', 'dev']
    regions = ['us-east-1', 'us-west-2', 'eu-west-1']

    records: List[BillingRecord] = []
    start = date(2024, 1, 1)

    for day_offset in range(days):
        period_start = start + timedelta(days=day_offset)
        period_end   = period_start + timedelta(days=1)

        for svc_name, svc_cat, base_cost, has_ri in services:
            team   = random.choice(teams)
            env    = random.choices(envs, weights=[0.6, 0.25, 0.15])[0]
            region = random.choice(regions)

            # Weekend effect — lower compute on weekends
            day_of_week = period_start.weekday()
            weekend_factor = 0.4 if day_of_week >= 5 and svc_cat == 'Compute' else 1.0

            list_cost = round(base_cost * random.uniform(0.8, 1.3) * weekend_factor, 4)

            # RI/SP discount applies to on-demand-eligible services
            if has_ri and env == 'prod':
                discount = random.uniform(0.30, 0.72)  # 30-72% discount
                billed_cost = round(list_cost * (1 - discount), 4)
            else:
                billed_cost = list_cost

            effective_cost = round(billed_cost * random.uniform(0.95, 1.05), 4)

            resource_id = f'arn:aws:{svc_name.lower()}:{region}:123456:{team}-{env}-001'

            records.append(BillingRecord(
                billing_period_start=period_start,
                billing_period_end=period_end,
                service_name=svc_name,
                service_category=svc_cat,
                resource_id=resource_id,
                region_name=region,
                billed_cost=billed_cost,
                list_cost=list_cost,
                effective_cost=effective_cost,
                provider_name='AWS',
                tags={
                    'team':        team,
                    'environment': env,
                    'cost-center': f'CC-{1000 + teams.index(team) * 100}',
                },
            ))

    return records


records = generate_billing_dataset(days=30)
print(f'Generated {len(records)} billing records')
print(f'Sample record:')
r = records[0]
print(f'  {r.billing_period_start}  {r.service_name:12s}  '
      f'list=${r.list_cost:.4f}  billed=${r.billed_cost:.4f}  discount={r.discount_pct:.1f}%')
```

```python
# Cell 2: Billing aggregator — cost by service, team, environment
from collections import defaultdict
from typing import Callable, Tuple


def aggregate_costs(
    records: List[BillingRecord],
    group_by: Callable[[BillingRecord], str],
) -> Dict[str, float]:
    """Aggregate billed_cost grouped by an arbitrary key function."""
    totals: Dict[str, float] = defaultdict(float)
    for r in records:
        totals[group_by(r)] += r.billed_cost
    return dict(sorted(totals.items(), key=lambda x: x[1], reverse=True))


def print_cost_table(title: str, data: Dict[str, float], total: float) -> None:
    print(f'\n=== {title} ===')
    print(f'  {"Group":<20s}  {"Cost (USD)":>12s}  {"Share":>8s}')
    print('  ' + '-' * 44)
    for group, cost in data.items():
        share = cost / total * 100 if total > 0 else 0
        print(f'  {group:<20s}  ${cost:>11.2f}  {share:>7.1f}%')
    print(f'  {"TOTAL":<20s}  ${total:>11.2f}  {100.0:>7.1f}%')


total_billed = sum(r.billed_cost for r in records)
total_list   = sum(r.list_cost   for r in records)
total_savings = total_list - total_billed

print(f'30-Day Cloud Bill Summary')
print(f'  On-demand (list):  ${total_list:.2f}')
print(f'  Actual billed:     ${total_billed:.2f}')
print(f'  Commitment savings:${total_savings:.2f}  ({total_savings/total_list*100:.1f}%)')

by_service = aggregate_costs(records, lambda r: r.service_name)
by_team    = aggregate_costs(records, lambda r: r.tags.get('team', 'untagged'))
by_env     = aggregate_costs(records, lambda r: r.tags.get('environment', 'untagged'))
by_category= aggregate_costs(records, lambda r: r.service_category)

print_cost_table('By Service',     by_service,  total_billed)
print_cost_table('By Team',        by_team,     total_billed)
print_cost_table('By Environment', by_env,      total_billed)
print_cost_table('By Category',    by_category, total_billed)
```

## Key Takeaways

- The FinOps cycle (Inform → Optimize → Operate) is continuous, not a one-time project
- FOCUS standardizes billing columns across providers, enabling multi-cloud analytics
- Tags are the foundation of cost allocation — without them, showback is impossible
- Unit economics (CPAU, cost per request) links cloud spend to business outcomes
- Commitment discounts (RI/SP) can reduce on-demand costs by 30–72%
