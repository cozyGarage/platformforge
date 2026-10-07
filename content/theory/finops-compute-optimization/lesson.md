Master the three commitment discount models (Reserved Instances, Savings Plans, Spot), understand rightsizing methodology, detect idle resources — then build a break-even optimizer that recommends the optimal commitment strategy for a given workload.

## Commitment Discount Models

| Model | Discount vs On-Demand | Flexibility | Commitment | Best For |
|---|---|---|---|---|
| **On-Demand** | 0% | Maximum | None | Unpredictable, short-lived |
| **Reserved Instance (RI)** | 30–72% | Locked to instance type/region | 1 or 3 years | Steady-state, predictable |
| **Savings Plans (SP)** | 20–66% | Flexible (instance family or compute) | 1 or 3 years | Variable mix, new instance types |
| **Spot Instances** | 70–90% | Can be interrupted with 2-min notice | None | Fault-tolerant batch, ML training |

### Reserved Instance Types

| RI Type | Scope | Flexibility |
|---|---|---|
| Standard RI | Specific instance type + region | No change after purchase |
| Convertible RI | Instance family + region | Can exchange for different type |
| Zonal RI | Specific AZ | Capacity reservation guaranteed |

### Savings Plans Types

| SP Type | Covers | Discount |
|---|---|---|
| Compute SP | EC2 + Lambda + Fargate (any region/OS/size) | Up to 66% |
| EC2 Instance SP | Specific EC2 instance family in one region | Up to 72% |
| SageMaker SP | SageMaker training + inference | Up to 64% |

## Rightsizing

Rightsizing matches instance size to actual workload requirements:

| Signal | Interpretation | Action |
|---|---|---|
| CPU avg < 10%, max < 30% | Over-provisioned for compute | Downsize or use Graviton |
| Memory avg < 20%, max < 40% | Over-provisioned for memory | Downsize instance family |
| Network throughput < 10% of capacity | Over-provisioned for network | Smaller instance |
| CPU avg > 80% sustained | Under-provisioned (performance risk) | Upsize |

### Idle Resource Detection

| Resource | Idle Signal |
|---|---|
| EC2 instance | CPU < 5% for 14 consecutive days |
| RDS instance | Connection count = 0 for 7 days |
| EBS volume | Not attached to any instance |
| Elastic IP | Not associated with running instance |
| Load balancer | Zero requests for 7 days |

```python
# Cell 1: RI / Savings Plan break-even calculator
from dataclasses import dataclass
from typing import Dict, List, Tuple
import math


@dataclass
class CommitmentOption:
    name: str
    hourly_rate: float       # effective hourly cost with commitment
    upfront_cost: float      # total upfront payment (0 for No Upfront)
    term_months: int
    commitment_type: str     # 'ri_standard', 'ri_convertible', 'savings_plan', 'on_demand', 'spot'
    flexibility: str         # 'none', 'low', 'high', 'maximum'
    interruption_risk: bool  # True for Spot

    @property
    def total_cost(self) -> float:
        """Total cost over full term including upfront."""
        return self.upfront_cost + (self.hourly_rate * 730 * self.term_months)


# AWS pricing for m5.xlarge (4 vCPU, 16 GB) — approximate 2024 values
M5_XLARGE_OPTIONS: List[CommitmentOption] = [
    CommitmentOption('On-Demand',                  0.192, 0,      1,  'on_demand',      'maximum', False),
    CommitmentOption('Spot',                        0.040, 0,      1,  'spot',           'maximum', True),
    CommitmentOption('1yr No Upfront RI',           0.122, 0,      12, 'ri_standard',    'none',    False),
    CommitmentOption('1yr Partial Upfront RI',      0.109, 480,    12, 'ri_standard',    'none',    False),
    CommitmentOption('1yr All Upfront RI',          0.101, 889,    12, 'ri_standard',    'none',    False),
    CommitmentOption('3yr No Upfront RI',           0.075, 0,      36, 'ri_standard',    'none',    False),
    CommitmentOption('3yr All Upfront RI',          0.067, 1771,   36, 'ri_standard',    'none',    False),
    CommitmentOption('1yr Compute Savings Plan',    0.128, 0,      12, 'savings_plan',   'high',    False),
    CommitmentOption('3yr Compute Savings Plan',    0.082, 0,      36, 'savings_plan',   'high',    False),
    CommitmentOption('1yr EC2 Instance SP',         0.115, 0,      12, 'savings_plan',   'low',     False),
    CommitmentOption('3yr EC2 Instance SP',         0.071, 0,      36, 'savings_plan',   'low',     False),
]


def break_even_analysis(
    options: List[CommitmentOption],
    baseline: CommitmentOption,
    usage_hours_per_month: float = 730,  # 730h = 100% utilization
) -> None:
    """Compare all commitment options against baseline (on-demand) and compute savings."""
    baseline_monthly = baseline.hourly_rate * usage_hours_per_month

    print(f'=== Commitment Discount Analysis: m5.xlarge ===')
    print(f'Baseline (On-Demand): ${baseline_monthly:.2f}/month  (${baseline.hourly_rate:.3f}/hr)')
    print(f'Usage: {usage_hours_per_month:.0f} hours/month')
    print()
    print(f'  {"Option":<32s}  {"$/mo":>8s}  {"Savings":>8s}  {"Term":>6s}  {"Break-even":>12s}  {"Flexibility":<12s}')
    print('  ' + '-' * 90)

    for opt in sorted(options, key=lambda o: o.hourly_rate * usage_hours_per_month + o.upfront_cost / max(o.term_months, 1)):
        if opt.name == baseline.name:
            continue

        monthly_cost  = opt.hourly_rate * usage_hours_per_month
        monthly_savings = baseline_monthly - monthly_cost
        savings_pct = monthly_savings / baseline_monthly * 100

        if opt.upfront_cost > 0 and monthly_savings > 0:
            break_even_months = opt.upfront_cost / monthly_savings
            be_str = f'{break_even_months:.1f} months'
        elif opt.upfront_cost == 0:
            be_str = 'immediate'
        else:
            be_str = 'never'

        interrupt = ' ⚠ interrupt' if opt.interruption_risk else ''
        print(f'  {opt.name:<32s}  ${monthly_cost:>7.2f}  '
              f'{savings_pct:>7.1f}%  {opt.term_months:>4}mo  '
              f'{be_str:>12s}  {opt.flexibility:<12s}{interrupt}')


on_demand = next(o for o in M5_XLARGE_OPTIONS if o.commitment_type == 'on_demand')
break_even_analysis(M5_XLARGE_OPTIONS, on_demand, usage_hours_per_month=730)
```

```python
# Cell 2: Rightsizing recommendation engine + idle resource detector
from dataclasses import dataclass, field
from typing import Dict, List, Optional
import statistics


@dataclass
class InstanceMetrics:
    instance_id: str
    instance_type: str
    vcpus: int
    memory_gb: float
    monthly_cost: float
    cpu_avg_pct: float       # 14-day average CPU utilization
    cpu_max_pct: float       # 14-day peak CPU utilization
    memory_avg_pct: float
    memory_max_pct: float
    network_avg_mbps: float
    connection_count: int    # For RDS: daily average connections
    service: str             # 'EC2', 'RDS'


@dataclass
class RightsizingRecommendation:
    instance: InstanceMetrics
    action: str              # 'terminate', 'downsize', 'keep', 'upsize'
    reason: str
    estimated_monthly_savings: float
    recommended_type: Optional[str] = None


# Simplified instance family sizing (each step is ~50% of resources)
EC2_FAMILIES: Dict[str, List[str]] = {
    'm5': ['m5.large', 'm5.xlarge', 'm5.2xlarge', 'm5.4xlarge', 'm5.8xlarge'],
    'r5': ['r5.large', 'r5.xlarge', 'r5.2xlarge', 'r5.4xlarge'],
    'c5': ['c5.large', 'c5.xlarge', 'c5.2xlarge', 'c5.4xlarge'],
}
HOURLY_COSTS: Dict[str, float] = {
    'm5.large': 0.096, 'm5.xlarge': 0.192, 'm5.2xlarge': 0.384, 'm5.4xlarge': 0.768,
    'r5.large': 0.126, 'r5.xlarge': 0.252, 'r5.2xlarge': 0.504,
    'c5.large': 0.085, 'c5.xlarge': 0.170, 'c5.2xlarge': 0.340,
}


def recommend_rightsizing(metrics: InstanceMetrics) -> RightsizingRecommendation:
    """Produce a rightsizing or termination recommendation based on utilization metrics."""

    # Idle detection
    if metrics.service == 'EC2' and metrics.cpu_avg_pct < 5.0 and metrics.cpu_max_pct < 15.0:
        return RightsizingRecommendation(
            instance=metrics, action='terminate',
            reason=f'CPU avg={metrics.cpu_avg_pct:.1f}%, max={metrics.cpu_max_pct:.1f}% — idle for 14 days',
            estimated_monthly_savings=metrics.monthly_cost,
        )

    if metrics.service == 'RDS' and metrics.connection_count == 0:
        return RightsizingRecommendation(
            instance=metrics, action='terminate',
            reason='Zero database connections for 7 days — likely orphaned',
            estimated_monthly_savings=metrics.monthly_cost,
        )

    # Over-provisioned: recommend downsize
    if metrics.cpu_avg_pct < 20.0 and metrics.memory_avg_pct < 30.0:
        # Find next smaller instance in the same family
        family = metrics.instance_type.rsplit('.', 1)[0]
        sizes = EC2_FAMILIES.get(family, [])
        current_idx = next((i for i, s in enumerate(sizes) if s == metrics.instance_type), -1)

        if current_idx > 0:
            smaller = sizes[current_idx - 1]
            new_cost = HOURLY_COSTS.get(smaller, 0) * 730
            savings = metrics.monthly_cost - new_cost
            return RightsizingRecommendation(
                instance=metrics, action='downsize',
                reason=f'CPU avg={metrics.cpu_avg_pct:.1f}%, mem avg={metrics.memory_avg_pct:.1f}% — over-provisioned',
                estimated_monthly_savings=round(savings, 2),
                recommended_type=smaller,
            )

    # Under-provisioned: warn
    if metrics.cpu_avg_pct > 80.0 or metrics.memory_avg_pct > 85.0:
        return RightsizingRecommendation(
            instance=metrics, action='upsize',
            reason=f'CPU avg={metrics.cpu_avg_pct:.1f}%, mem avg={metrics.memory_avg_pct:.1f}% — performance risk',
            estimated_monthly_savings=-metrics.monthly_cost * 0.5,  # negative = additional cost
        )

    return RightsizingRecommendation(
        instance=metrics, action='keep',
        reason='Utilization within healthy range',
        estimated_monthly_savings=0.0,
    )


# Sample fleet
fleet = [
    InstanceMetrics('i-aaa111', 'm5.2xlarge', 8,  32, 280, cpu_avg_pct=3.2,  cpu_max_pct=9.8,  memory_avg_pct=8.0,  memory_max_pct=12.0, network_avg_mbps=10,  connection_count=0,  service='EC2'),
    InstanceMetrics('i-bbb222', 'm5.xlarge',  4,  16, 140, cpu_avg_pct=18.0, cpu_max_pct=45.0, memory_avg_pct=22.0, memory_max_pct=55.0, network_avg_mbps=50,  connection_count=0,  service='EC2'),
    InstanceMetrics('i-ccc333', 'r5.2xlarge', 8,  64, 368, cpu_avg_pct=72.0, cpu_max_pct=95.0, memory_avg_pct=88.0, memory_max_pct=97.0, network_avg_mbps=200, connection_count=0,  service='EC2'),
    InstanceMetrics('db-prod1', 'r5.xlarge',  4,  32, 184, cpu_avg_pct=35.0, cpu_max_pct=60.0, memory_avg_pct=45.0, memory_max_pct=70.0, network_avg_mbps=80,  connection_count=120, service='RDS'),
    InstanceMetrics('db-stg01', 'r5.xlarge',  4,  32, 184, cpu_avg_pct=2.0,  cpu_max_pct=5.0,  memory_avg_pct=10.0, memory_max_pct=15.0, network_avg_mbps=5,   connection_count=0,   service='RDS'),
    InstanceMetrics('i-ddd444', 'm5.4xlarge', 16, 64, 560, cpu_avg_pct=12.0, cpu_max_pct=28.0, memory_avg_pct=15.0, memory_max_pct=35.0, network_avg_mbps=40,  connection_count=0,  service='EC2'),
]

recommendations = [recommend_rightsizing(m) for m in fleet]

print('=== Rightsizing Recommendations ===')
print(f'  {"Instance":<12s}  {"Type":<14s}  {"Action":<10s}  {"$/mo Savings":>12s}  Reason')
print('  ' + '-' * 80)
total_savings = 0.0
for rec in recommendations:
    m = rec.instance
    new_type = f' → {rec.recommended_type}' if rec.recommended_type else ''
    savings_str = f'${rec.estimated_monthly_savings:.2f}' if rec.estimated_monthly_savings != 0 else '$0'
    print(f'  {m.instance_id:<12s}  {m.instance_type + new_type:<22s}  '
          f'{rec.action.upper():<10s}  {savings_str:>10s}  {rec.reason}')
    if rec.estimated_monthly_savings > 0:
        total_savings += rec.estimated_monthly_savings

print(f'\n  Total monthly savings opportunity: ${total_savings:.2f}/mo  (${total_savings*12:.0f}/yr)')
```

## Commitment Strategy Decision Tree

```
Workload type?
├── Fault-tolerant batch / ML training  →  Spot (70–90% off)
├── Steady-state, same instance type    →  Standard RI (up to 72% off)
├── Steady-state, may change type       →  Savings Plan (up to 66% off)
└── Unpredictable, bursty               →  On-Demand (no commitment)

Coverage target:
  - RI/SP cover baseline (p10 of hourly usage)
  - On-Demand covers burst above baseline
  - Spot covers opportunistic batch above baseline
```

## Key Takeaways

- Spot instances offer the largest discount but require interruption handling (SQS, checkpointing)
- Savings Plans are generally preferred over RIs due to flexibility across instance types
- Cover only your baseline usage with commitments — not peak — to avoid wasted reservations
- Idle EC2 (CPU < 5% avg) and zero-connection RDS are prime termination candidates
- Rightsizing before purchasing RIs avoids locking in discount on over-provisioned resources
