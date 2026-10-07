**Scenario**: You are the lead FinOps engineer at Nexlify, a B2B SaaS company with $180K/month in cloud spend. The CFO flags that cloud costs grew 60% in 6 months while revenue grew only 20%. Your task: investigate, quantify waste, and produce an optimization roadmap with projected savings.

## Company Background

| Attribute | Value |
|---|---|
| Monthly cloud spend | $180,000 |
| Monthly revenue | $1,200,000 |
| Monthly active users | 42,000 |
| Cloud as % of revenue | 15% (target: < 8%) |
| CPAU | $4.29 (target: < $2.00) |
| Engineering team size | 45 engineers, 8 teams |
| Deployment platform | AWS (EC2, EKS, RDS, S3, CloudFront) |
| Tag compliance | 61% (many untagged resources) |
| RI/SP coverage | 18% (almost all on-demand) |

## Investigation Approach

1. **Waste Analysis** — identify idle, over-provisioned, and unoptimized resources
2. **Optimization Plan** — prioritize by ROI, estimate savings, assign owners
3. **90-Day Roadmap** — phase the interventions by effort and impact

```python
# Cell 1: Waste Analysis — identify all waste categories across the Nexlify cloud account
from dataclasses import dataclass, field
from typing import Dict, List, Optional
from collections import defaultdict


@dataclass
class WasteItem:
    category: str           # 'idle_compute', 'oversized', 'unoptimized_storage', 'no_commitment', 'egress'
    resource_id: str
    resource_type: str
    team: str
    environment: str
    monthly_waste: float    # estimated wasteful spend in USD/month
    monthly_total: float    # total monthly cost of this resource
    waste_reason: str
    effort_to_fix: str      # 'low', 'medium', 'high'
    fix_action: str

    @property
    def waste_pct(self) -> float:
        return (self.monthly_waste / self.monthly_total * 100) if self.monthly_total > 0 else 0


# Full waste inventory across Nexlify's AWS account
waste_inventory: List[WasteItem] = [

    # === IDLE COMPUTE ===
    WasteItem('idle_compute', 'i-staging-web-01', 'EC2 m5.2xlarge', 'product', 'staging',
              monthly_waste=280, monthly_total=280,
              waste_reason='CPU avg 1.2%, zero traffic — never decommissioned after migration',
              effort_to_fix='low', fix_action='Terminate instance; save snapshot'),

    WasteItem('idle_compute', 'i-staging-web-02', 'EC2 m5.2xlarge', 'product', 'staging',
              monthly_waste=280, monthly_total=280,
              waste_reason='Duplicate of staging-web-01, both serving zero requests',
              effort_to_fix='low', fix_action='Terminate instance'),

    WasteItem('idle_compute', 'db-analytics-dev', 'RDS r5.2xlarge', 'data', 'dev',
              monthly_waste=550, monthly_total=550,
              waste_reason='Zero connections for 21 days — analyst moved to Redshift Serverless',
              effort_to_fix='low', fix_action='Create final snapshot and terminate'),

    WasteItem('idle_compute', 'elb-old-api', 'Application Load Balancer', 'platform', 'prod',
              monthly_waste=22, monthly_total=22,
              waste_reason='No targets registered, zero requests for 90 days',
              effort_to_fix='low', fix_action='Delete ALB'),

    WasteItem('idle_compute', 'cache-reports-stg', 'ElastiCache r5.large', 'product', 'staging',
              monthly_waste=140, monthly_total=140,
              waste_reason='Reports feature disabled in staging; cache unused',
              effort_to_fix='low', fix_action='Delete cluster; re-provision on demand'),

    WasteItem('idle_compute', 'nat-gw-dev', 'NAT Gateway', 'infra', 'dev',
              monthly_waste=32, monthly_total=32,
              waste_reason='Dev environment uses NAT GW but traffic is < 1 GB/month',
              effort_to_fix='medium', fix_action='Replace with NAT instance or route through VPN'),

    # === OVERSIZED INSTANCES ===
    WasteItem('oversized', 'i-api-server-prod-01', 'EC2 m5.4xlarge', 'product', 'prod',
              monthly_waste=280, monthly_total=560,
              waste_reason='CPU avg 12%, memory avg 18% — 4x over-provisioned vs p95 load',
              effort_to_fix='medium', fix_action='Rightsize to m5.xlarge after load testing'),

    WasteItem('oversized', 'i-api-server-prod-02', 'EC2 m5.4xlarge', 'product', 'prod',
              monthly_waste=280, monthly_total=560,
              waste_reason='CPU avg 14%, memory avg 20% — provisioned for "future growth" 18 months ago',
              effort_to_fix='medium', fix_action='Rightsize to m5.xlarge'),

    WasteItem('oversized', 'db-prod-primary', 'RDS r5.4xlarge', 'product', 'prod',
              monthly_waste=550, monthly_total=1100,
              waste_reason='CPU avg 8%, connections avg 12 — sized for peak never reached',
              effort_to_fix='high', fix_action='Rightsize to r5.xlarge with maintenance window'),

    WasteItem('oversized', 'eks-nodegroup-general', 'EKS m5.2xlarge nodes (8 nodes)', 'platform', 'prod',
              monthly_waste=890, monthly_total=2240,
              waste_reason='Node efficiency 34% CPU / 28% memory — over-provisioned node pool',
              effort_to_fix='medium', fix_action='Enable Cluster Autoscaler; rightsize to m5.xlarge nodes'),

    WasteItem('oversized', 'i-ml-training-fleet', 'EC2 p3.8xlarge (3 instances, always-on)', 'ml', 'prod',
              monthly_waste=3200, monthly_total=4800,
              waste_reason='Training runs only 8h/day; instances idle 16h/day with no auto-stop',
              effort_to_fix='medium', fix_action='Migrate to Spot + SageMaker managed training jobs'),

    # === UNOPTIMIZED STORAGE ===
    WasteItem('unoptimized_storage', 's3-raw-data-lake', 'S3 Standard (48 TB)', 'data', 'prod',
              monthly_waste=736, monthly_total=1104,
              waste_reason='48 TB raw data in S3 Standard; data > 30 days old never accessed',
              effort_to_fix='low', fix_action='Add lifecycle policy: IA at 30d, Glacier at 90d'),

    WasteItem('unoptimized_storage', 's3-backups-prod', 'S3 Standard (120 TB)', 'infra', 'prod',
              monthly_waste=1932, monthly_total=2760,
              waste_reason='Database backups stored in Standard; all > 7 days old, kept for 7 years',
              effort_to_fix='low', fix_action='Lifecycle to Glacier Deep Archive at 7d (70% savings)'),

    WasteItem('unoptimized_storage', 'ebs-snapshots-dev', 'EBS Snapshots (8 TB)', 'product', 'dev',
              monthly_waste=320, monthly_total=320,
              waste_reason='Dev snapshots never deleted; 180 days of daily snapshots retained',
              effort_to_fix='low', fix_action='Retention policy: keep 7 days for dev, 30 for staging'),

    # === NO COMMITMENT DISCOUNTS ===
    WasteItem('no_commitment', 'compute-on-demand-all', 'All EC2/EKS (82% on-demand)', 'platform', 'prod',
              monthly_waste=14400, monthly_total=48000,
              waste_reason='Only 18% RI/SP coverage; stable baseline workload is 100% on-demand',
              effort_to_fix='low', fix_action='Purchase 1yr Compute Savings Plans for 60% baseline coverage'),

    # === DATA TRANSFER / EGRESS ===
    WasteItem('egress', 'direct-s3-api-egress', 'S3 Direct Egress (8 TB/mo)', 'product', 'prod',
              monthly_waste=540, monthly_total=720,
              waste_reason='Static assets served directly from S3 at $0.09/GB; no CloudFront',
              effort_to_fix='medium', fix_action='Route through CloudFront (85% cache hit → ~75% savings)'),

    WasteItem('egress', 'cross-region-replication', 'S3 Cross-Region Replication (15 TB/mo)', 'data', 'prod',
              monthly_waste=225, monthly_total=300,
              waste_reason='Replicating all raw data to eu-west-1; EU ops only need processed data',
              effort_to_fix='medium', fix_action='Replicate only processed tier; reduce 75% of cross-region egress'),
]


def print_waste_analysis(inventory: List[WasteItem]) -> None:
    """Print waste analysis by category with totals."""
    total_waste = sum(w.monthly_waste for w in inventory)
    total_spend = 180_000  # Nexlify total monthly spend
    waste_pct   = total_waste / total_spend * 100

    print(f'=== Nexlify Cloud Waste Analysis ===')
    print(f'  Total monthly cloud spend:  ${total_spend:>10,.0f}')
    print(f'  Identified monthly waste:   ${total_waste:>10,.0f}')
    print(f'  Waste percentage:           {waste_pct:>9.1f}%')
    print(f'  Annual waste (if unchanged):${total_waste*12:>10,.0f}')
    print()

    # By category
    by_cat: Dict[str, List[WasteItem]] = defaultdict(list)
    for w in inventory:
        by_cat[w.category].append(w)

    cat_labels = {
        'idle_compute':         'Idle Compute',
        'oversized':            'Oversized Resources',
        'unoptimized_storage':  'Unoptimized Storage',
        'no_commitment':        'No Commitment Discounts',
        'egress':               'Data Transfer / Egress',
    }

    print(f'  {"Category":<30s}  {"Waste/mo":>10s}  {"% of Waste":>10s}  Items')
    print('  ' + '-' * 62)
    for cat, items in sorted(by_cat.items(), key=lambda x: sum(i.monthly_waste for i in x[1]), reverse=True):
        cat_waste = sum(i.monthly_waste for i in items)
        cat_pct   = cat_waste / total_waste * 100
        print(f'  {cat_labels.get(cat, cat):<30s}  ${cat_waste:>9,.0f}  {cat_pct:>9.1f}%  {len(items)}')
    print()

    # Top 5 individual waste items
    print('  Top 5 Waste Items by Monthly Cost:')
    for w in sorted(inventory, key=lambda x: x.monthly_waste, reverse=True)[:5]:
        print(f'    ${w.monthly_waste:>6,.0f}/mo  [{w.effort_to_fix.upper():6s}]  '
              f'{w.resource_id}  ({w.resource_type})')
        print(f'               → {w.waste_reason[:80]}')

    # By effort
    print()
    print('  Waste by Remediation Effort:')
    for effort in ['low', 'medium', 'high']:
        items = [w for w in inventory if w.effort_to_fix == effort]
        effort_waste = sum(w.monthly_waste for w in items)
        print(f'    {effort.upper():6s}: {len(items):2d} items  ${effort_waste:>8,.0f}/mo  '
              f'(${effort_waste*12:>10,.0f}/yr)')


print_waste_analysis(waste_inventory)
```

```python
# Cell 2: Optimization plan with 90-day roadmap and projected savings
from dataclasses import dataclass
from typing import List, Dict


@dataclass
class OptimizationAction:
    phase: int              # 1=immediate, 2=30-day, 3=60-90-day
    category: str
    action: str
    owner: str
    monthly_savings: float
    one_time_cost: float    # implementation cost (engineer hours × rate)
    payback_months: float
    risk: str               # 'none', 'low', 'medium'

    @property
    def annual_net_savings(self) -> float:
        return self.monthly_savings * 12 - self.one_time_cost

    @property
    def roi_pct(self) -> float:
        if self.one_time_cost == 0:
            return float('inf')
        return self.annual_net_savings / self.one_time_cost * 100


optimization_plan: List[OptimizationAction] = [

    # Phase 1: Immediate (0–2 weeks) — no service impact, low effort
    OptimizationAction(1, 'Idle Cleanup', 'Terminate 5 idle compute resources (staging EC2x2, RDS, ELB, ElastiCache)',
                       'Platform Team', monthly_savings=1074, one_time_cost=800, payback_months=0.7, risk='none'),

    OptimizationAction(1, 'Storage Lifecycle', 'Add S3 lifecycle policy to raw data lake (48 TB → IA+Glacier)',
                       'Data Team', monthly_savings=736, one_time_cost=400, payback_months=0.5, risk='none'),

    OptimizationAction(1, 'Storage Lifecycle', 'Glacier Deep Archive for backup bucket (120 TB)',
                       'Infra Team', monthly_savings=1932, one_time_cost=400, payback_months=0.2, risk='none'),

    OptimizationAction(1, 'Storage Cleanup', 'Delete stale EBS snapshots (retain 7d dev, 30d staging)',
                       'Product Team', monthly_savings=320, one_time_cost=200, payback_months=0.6, risk='none'),

    OptimizationAction(1, 'Tagging', 'Tag all untagged resources via Terraform import + tag policy',
                       'Platform Team', monthly_savings=0, one_time_cost=2400, payback_months=0, risk='none'),
    # ^ tagging enables chargeback → indirect savings not counted here

    # Phase 2: 30-Day — requires testing/validation, medium effort
    OptimizationAction(2, 'Commitment Discounts', 'Purchase 1yr Compute Savings Plans (60% baseline coverage)',
                       'FinOps + Finance', monthly_savings=14400, one_time_cost=1200, payback_months=0.1, risk='low'),
    # Savings Plans require upfront planning but are financial instrument, not technical change

    OptimizationAction(2, 'Rightsizing', 'Downsize 2x m5.4xlarge API servers to m5.xlarge (load test first)',
                       'Product Team', monthly_savings=560, one_time_cost=2400, payback_months=4.3, risk='low'),

    OptimizationAction(2, 'CloudFront', 'Route S3 static asset delivery through CloudFront CDN',
                       'Product Team', monthly_savings=540, one_time_cost=1600, payback_months=3.0, risk='low'),

    OptimizationAction(2, 'ML Optimization', 'Migrate ML training to Spot + SageMaker (8h/day usage pattern)',
                       'ML Team', monthly_savings=3200, one_time_cost=4800, payback_months=1.5, risk='medium'),

    OptimizationAction(2, 'Egress', 'Reduce cross-region replication to processed data only',
                       'Data Team', monthly_savings=225, one_time_cost=800, payback_months=3.6, risk='low'),

    # Phase 3: 60–90 Day — higher complexity, maintenance windows needed
    OptimizationAction(3, 'Rightsizing', 'Rightsize RDS r5.4xlarge → r5.xlarge (requires maintenance window)',
                       'Product Team', monthly_savings=550, one_time_cost=3200, payback_months=5.8, risk='medium'),

    OptimizationAction(3, 'K8s Efficiency', 'Enable Cluster Autoscaler + rightsize EKS nodegroup to m5.xlarge',
                       'Platform Team', monthly_savings=890, one_time_cost=3200, payback_months=3.6, risk='medium'),

    OptimizationAction(3, 'Dev Environments', 'Scale-to-zero dev/staging environments (nights+weekends via scheduler)',
                       'Platform Team', monthly_savings=1800, one_time_cost=4000, payback_months=2.2, risk='low'),
    # ^ estimated 65% reduction in non-prod compute
]


def print_optimization_roadmap(actions: List[OptimizationAction]) -> None:
    total_spend = 180_000
    phase_labels = {1: 'Phase 1 — Immediate (0–2 weeks)', 2: 'Phase 2 — 30 Days', 3: 'Phase 3 — 60–90 Days'}

    print('=== Nexlify Optimization Roadmap ===')
    print()

    cumulative_savings = 0.0
    total_one_time = sum(a.one_time_cost for a in actions)
    total_monthly  = sum(a.monthly_savings for a in actions)

    for phase in [1, 2, 3]:
        phase_actions = [a for a in actions if a.phase == phase]
        phase_monthly = sum(a.monthly_savings for a in phase_actions)
        phase_cost    = sum(a.one_time_cost   for a in phase_actions)
        cumulative_savings += phase_monthly

        print(f'  {phase_labels[phase]}')
        print(f'  Phase savings: ${phase_monthly:,.0f}/mo  Implementation cost: ${phase_cost:,.0f}')
        print(f'  {"Action":<55s}  {"Owner":<15s}  {"$/mo":>8s}  {"Risk":<8s}  {"Payback"}')
        print('  ' + '-' * 105)
        for a in sorted(phase_actions, key=lambda x: x.monthly_savings, reverse=True):
            payback_str = f'{a.payback_months:.1f}mo' if a.payback_months > 0 else 'immediate'
            print(f'  {a.action[:55]:<55s}  {a.owner:<15s}  ${a.monthly_savings:>7,.0f}  {a.risk:<8s}  {payback_str}')
        print()

    # Summary
    new_monthly = total_spend - total_monthly
    new_pct_rev = new_monthly / 1_200_000 * 100  # assume revenue stays same

    print('  === Projected Impact Summary ===')
    print(f'  Current monthly spend:          ${total_spend:>10,.0f}  ({total_spend/1_200_000*100:.1f}% of revenue)')
    print(f'  Total monthly savings:          ${total_monthly:>10,.0f}')
    print(f'  Projected monthly spend:        ${new_monthly:>10,.0f}  ({new_pct_rev:.1f}% of revenue)')
    print(f'  Reduction:                      {total_monthly/total_spend*100:>9.1f}%')
    print(f'  Total implementation cost:      ${total_one_time:>10,.0f}')
    print(f'  Annual net savings (yr 1):      ${total_monthly*12 - total_one_time:>10,.0f}')
    print(f'  ROI:                            {(total_monthly*12 - total_one_time)/total_one_time*100:>9.0f}%')
    print()
    print('  === KPI Improvement ===')
    old_cpau = total_spend / 42_000
    new_cpau = new_monthly / 42_000
    print(f'  CPAU: ${old_cpau:.2f} → ${new_cpau:.2f}  (target: < $2.00)')
    print(f'  Cloud % of revenue: {total_spend/1_200_000*100:.1f}% → {new_pct_rev:.1f}%  (target: < 8%)')
    print(f'  RI/SP Coverage: 18% → ~65%  (after Phase 2 Savings Plans purchase)')
    print(f'  Tag Compliance: 61% → ~95%  (after Phase 1 tagging sprint)')


print_optimization_roadmap(optimization_plan)
```

## Case Study Lessons

| Finding | Root Cause | Prevention |
|---|---|---|
| Idle EC2/RDS in staging | No lifecycle policy for non-prod | Auto-shutdown via scheduler; quarterly idle audit |
| Zero RI/SP coverage | Finance not engaged in commitment decisions | Monthly FinOps review with Finance |
| 40% over-provisioned compute | Dev sizing carried to prod | Load testing gating for prod deployments |
| 120 TB backups in Standard | No lifecycle policy on backup bucket | Tag-based lifecycle policies in Terraform modules |
| Static assets without CDN | Architecture review missed egress modeling | Add egress cost estimate to architecture checklist |
| ML training instances always-on | No cost visibility for ML team | ML team showback dashboard + Spot enforcement policy |

## Key Takeaways

- 40% waste in a growing SaaS is common without active FinOps practice
- Savings Plans (commitment discounts) are the single highest-impact lever: $14,400/mo savings here
- Storage lifecycle policies are zero-risk, high-ROI: implement as Terraform defaults
- Idle resource cleanup requires discovery tooling (AWS Cost Explorer, Trusted Advisor) run weekly
- Implementation prioritization: sort by (monthly_savings / one_time_cost) — maximize ROI first
- After optimization: cloud drops from 15% to ~6.3% of revenue — below the 8% target
