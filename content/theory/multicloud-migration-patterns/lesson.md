## 7.1  The 7 Rs of Cloud Migration

| Strategy | Also called | Effort | Risk | When to use |
|----------|------------|--------|------|-------------|
| **Retire** | Remove | None | None | Application is no longer needed |
| **Retain** | Revisit | None | None | Too risky to migrate now; revisit in 6-12 months |
| **Rehost** | Lift-and-shift | Low | Low | Speed is priority; optimise later |
| **Relocate** | Hypervisor lift | Low | Low | VMware-to-VMware Cloud migration |
| **Replatform** | Lift-and-tinker | Medium | Medium | Swap DB engine, containerise, minimal code changes |
| **Repurchase** | Drop-and-shop | Medium | Medium | Replace with SaaS (e.g. Salesforce, Workday) |
| **Rearchitect** | Refactor | High | High | Modernise to microservices, serverless, cloud-native |

## 7.2  Migration Execution Patterns

### Strangler Fig
Gradually replace a monolith by routing specific features/endpoints to new services:
```
                    ┌─────────────────────┐
Client → Proxy ─────┤                     ├── /api/orders  → [NEW: GCP microservice]
                    │   Strangler Facade   ├── /api/products → [NEW: AWS Lambda]
                    │   (API Gateway /    ├── /api/legacy   → [OLD: Monolith on-prem]
                    │    Load Balancer)   |
                    └─────────────────────┘
```
Over time, legacy routes shrink to zero and the monolith is retired.

### Parallel Run
Run old and new systems simultaneously; compare outputs before cutting over:
- Shadow traffic to new system; verify results match
- Gradually increase shadow traffic percentage
- Cut over when confidence threshold reached (e.g. 99.9% match)

### Blue-Green Cutover
Blue = current production. Green = new cloud target. Switch DNS/load balancer atomically.
- Instant rollback: flip DNS back to Blue
- Requires running two full environments simultaneously during transition

```python
# Migration wave planner
# Models application portfolio prioritisation and wave scheduling

from dataclasses import dataclass, field
from typing import Dict, List, Optional
from enum import Enum


class MigrationStrategy(str, Enum):
    RETIRE      = 'Retire'
    RETAIN      = 'Retain'
    REHOST      = 'Rehost'
    REPLATFORM  = 'Replatform'
    REARCHITECT = 'Rearchitect'


@dataclass
class Application:
    name: str
    business_criticality: int   # 1 (low) – 5 (critical)
    complexity: int             # 1 (simple) – 5 (complex)
    dependencies: List[str]     # names of apps this depends on
    strategy: MigrationStrategy
    effort_weeks: int
    team: str
    target_cloud: str


def dependency_score(app: Application, all_apps: Dict[str, 'Application']) -> int:
    """Count how many OTHER apps depend on this app (higher = migrate last)."""
    return sum(
        1 for a in all_apps.values()
        if app.name in a.dependencies and a.name != app.name
    )


def migration_priority(app: Application, all_apps: Dict[str, 'Application']) -> float:
    """
    Priority score: high business value, low complexity, few dependents → migrate early.
    Low score = higher priority (sort ascending).
    """
    if app.strategy in (MigrationStrategy.RETIRE, MigrationStrategy.RETAIN):
        return 999.0  # always last
    dep_score = dependency_score(app, all_apps)
    return (app.complexity * 2 + dep_score * 3) - app.business_criticality


def assign_waves(apps: List[Application], max_parallel_teams: int = 2) -> Dict[int, List[Application]]:
    """Assign applications to migration waves respecting dependencies."""
    app_map = {a.name: a for a in apps}
    sorted_apps = sorted(
        [a for a in apps if a.strategy not in (MigrationStrategy.RETIRE, MigrationStrategy.RETAIN)],
        key=lambda a: migration_priority(a, app_map),
    )

    waves: Dict[int, List[Application]] = {}
    migrated = set()
    wave = 1

    while sorted_apps:
        wave_apps = []
        remaining = []
        for app in sorted_apps:
            deps_migrated = all(d in migrated for d in app.dependencies)
            if deps_migrated and len(wave_apps) < max_parallel_teams:
                wave_apps.append(app)
                migrated.add(app.name)
            else:
                remaining.append(app)
        if not wave_apps:
            # Circular dependency or can't parallelise further — force next
            wave_apps = remaining[:1]
            migrated.add(wave_apps[0].name)
            remaining = remaining[1:]
        waves[wave] = wave_apps
        sorted_apps = remaining
        wave += 1

    return waves


# Sample application portfolio
portfolio = [
    Application('auth-service',     business_criticality=5, complexity=2, dependencies=[],
                strategy=MigrationStrategy.REPLATFORM, effort_weeks=4, team='platform', target_cloud='aws'),
    Application('user-api',         business_criticality=4, complexity=2, dependencies=['auth-service'],
                strategy=MigrationStrategy.REHOST, effort_weeks=3, team='backend', target_cloud='aws'),
    Application('ml-pipeline',      business_criticality=3, complexity=4, dependencies=['user-api'],
                strategy=MigrationStrategy.REARCHITECT, effort_weeks=12, team='ml', target_cloud='gcp'),
    Application('reporting-db',     business_criticality=3, complexity=3, dependencies=[],
                strategy=MigrationStrategy.REPLATFORM, effort_weeks=6, team='data', target_cloud='aws'),
    Application('analytics-ui',     business_criticality=2, complexity=1, dependencies=['reporting-db'],
                strategy=MigrationStrategy.REHOST, effort_weeks=2, team='frontend', target_cloud='azure'),
    Application('legacy-crm',       business_criticality=2, complexity=5, dependencies=[],
                strategy=MigrationStrategy.RETAIN, effort_weeks=0, team='none', target_cloud='on-prem'),
    Application('old-batch-system', business_criticality=1, complexity=1, dependencies=[],
                strategy=MigrationStrategy.RETIRE, effort_weeks=0, team='none', target_cloud='none'),
]

app_map = {a.name: a for a in portfolio}
waves   = assign_waves(portfolio, max_parallel_teams=2)

print('=== Migration Wave Plan ===')
total_weeks = 0
for wave_num, wave_apps in sorted(waves.items()):
    wave_duration = max(a.effort_weeks for a in wave_apps)
    total_weeks  += wave_duration
    print(f'\nWave {wave_num} ({wave_duration} weeks):')
    for app in wave_apps:
        deps_str = f' [deps: {app.dependencies}]' if app.dependencies else ''
        print(f'  • {app.name:<25} {app.strategy.value:<14} → {app.target_cloud:<8} team={app.team}{deps_str}')

# Retirements and retentions
special = [a for a in portfolio if a.strategy in (MigrationStrategy.RETIRE, MigrationStrategy.RETAIN)]
if special:
    print('\nNo-migration items:')
    for app in special:
        print(f'  • {app.name:<25} {app.strategy.value}')

print(f'\nTotal migration duration (serial waves): ~{total_weeks} weeks')
```

## 7.3  Migration Risk and Rollback

| Risk | Mitigation |
|------|------------|
| **Data loss during cutover** | Dual-write to old and new stores; validate checksums |
| **Performance regression** | Load test at 2× expected peak before cutover |
| **DNS propagation lag** | Set TTL to 30s 48 hours before cutover |
| **Dependency not migrated** | Dependency mapping + reverse dependency check in wave plan |
| **Credentials not ported** | Secret rotation during migration; use Secrets Manager, not env vars |
| **Rollback window** | Keep old environment live for 48-72 hours post-cutover |

**Definition of done for a migration wave:**
1. All smoke tests pass on new cloud
2. Synthetic monitoring active
3. Rollback procedure tested (not just written)
4. Cost baseline captured
5. Old environment decommission date scheduled

```python
# Strangler fig pattern simulator
# Models incremental traffic migration from monolith to cloud microservices

from dataclasses import dataclass, field
from typing import Dict, List, Tuple
import random


@dataclass
class RoutingRule:
    path_prefix: str
    target: str          # 'monolith' | 'new-service'
    weight: float        # 0.0-1.0 fraction to new service (rest to monolith)


@dataclass
class ServiceMetrics:
    name: str
    latency_p50_ms: float
    latency_p99_ms: float
    error_rate: float    # 0.0-1.0


MONOLITH_METRICS   = ServiceMetrics('monolith',     320, 1200, 0.008)
NEW_SERVICE_METRICS = ServiceMetrics('new-service',  45,   180, 0.001)


def route_request(path: str, rules: List[RoutingRule]) -> Tuple[str, ServiceMetrics]:
    """Route a request and return which system handled it."""
    for rule in rules:
        if path.startswith(rule.path_prefix):
            target_metrics = (
                NEW_SERVICE_METRICS if random.random() < rule.weight
                else MONOLITH_METRICS
            )
            return (rule.target if random.random() < rule.weight else 'monolith'), target_metrics
    return 'monolith', MONOLITH_METRICS


def simulate_migration_phase(
    phase_name: str,
    rules: List[RoutingRule],
    n_requests: int = 1000,
) -> None:
    random.seed(42)
    handled: Dict[str, int] = {'monolith': 0, 'new-service': 0}
    errors: Dict[str, int]  = {'monolith': 0, 'new-service': 0}
    latencies: Dict[str, List[float]] = {'monolith': [], 'new-service': []}

    test_paths = ['/api/orders', '/api/products', '/api/users', '/api/legacy']
    for i in range(n_requests):
        path = random.choice(test_paths)
        target, metrics = route_request(path, rules)
        handled[target] += 1
        # Simulate latency with jitter
        lat = random.gauss(metrics.latency_p50_ms, metrics.latency_p50_ms * 0.2)
        latencies[target].append(max(1, lat))
        if random.random() < metrics.error_rate:
            errors[target] += 1

    print(f'\n=== {phase_name} ===')
    for svc in ['monolith', 'new-service']:
        count = handled[svc]
        if count == 0:
            continue
        avg_lat = sum(latencies[svc]) / count
        err_pct = errors[svc] / count * 100
        pct = count / n_requests * 100
        print(f'  {svc:<15} {pct:>5.1f}% of traffic  avg latency={avg_lat:>6.0f}ms  errors={err_pct:.2f}%')


# Phase 1: 0% migrated (all on monolith)
simulate_migration_phase('Phase 1: Baseline (0% migrated)', [
    RoutingRule('/api/orders',   'new-service', weight=0.0),
    RoutingRule('/api/products', 'new-service', weight=0.0),
])

# Phase 2: /api/orders fully migrated, /api/products at 50%
simulate_migration_phase('Phase 2: Orders migrated, Products 50%', [
    RoutingRule('/api/orders',   'new-service', weight=1.0),
    RoutingRule('/api/products', 'new-service', weight=0.5),
])

# Phase 3: All migrated endpoints at 100%
simulate_migration_phase('Phase 3: Full cutover', [
    RoutingRule('/api/orders',   'new-service', weight=1.0),
    RoutingRule('/api/products', 'new-service', weight=1.0),
    RoutingRule('/api/users',    'new-service', weight=1.0),
])
```

## Summary

1. **Migrate easy wins first** — Low complexity, no dependents, low business criticality gives the team confidence before tackling critical workloads.
2. **Strangler fig is the default pattern** — It allows incremental migration with instant rollback at every step. Big-bang cutovers compound risk.
3. **Dependency mapping is prerequisite** — Migrating an app without migrating what it depends on leaves it stranded. Build the dependency graph before the wave plan.
4. **Rollback is part of the migration** — Test the rollback procedure on a low-criticality app before you need it on a critical one.

**Next →** [`08_exercises`](08_exercises.ipynb) — apply everything you've learned.
