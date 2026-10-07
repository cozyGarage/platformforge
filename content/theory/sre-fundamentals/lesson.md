Core vocabulary: **SLI → SLO → SLA → Error Budget → Toil**

## 1. SLI — Service Level Indicator

An SLI is a **quantitative measure** of some aspect of the service level being provided.

Common SLI categories:
- **Availability** — fraction of time the service is usable
- **Latency** — proportion of requests served within a threshold
- **Throughput** — rate of successful requests per second
- **Error rate** — fraction of requests that return errors
- **Freshness** — how recently data was updated (for data pipelines)

Formula (request-based SLI):
```
SLI = good_events / total_events
```

```python
from dataclasses import dataclass, field
from typing import List


@dataclass
class SLIWindow:
    """Tracks good/bad events over a rolling window."""
    name: str
    good_events: int = 0
    total_events: int = 0

    def record(self, *, good: bool) -> None:
        self.total_events += 1
        if good:
            self.good_events += 1

    @property
    def value(self) -> float:
        if self.total_events == 0:
            return 1.0
        return self.good_events / self.total_events

    @property
    def percentage(self) -> float:
        return self.value * 100


# Simulate a latency SLI (good = served within 200 ms)
import random
random.seed(42)

latency_sli = SLIWindow(name="p99_latency_under_200ms")
for _ in range(10_000):
    latency_ms = random.gauss(mu=120, sigma=50)
    latency_sli.record(good=latency_ms < 200)

print(f"SLI '{latency_sli.name}': {latency_sli.percentage:.3f}%")
print(f"  Good events : {latency_sli.good_events}")
print(f"  Total events: {latency_sli.total_events}")
```

## 2. SLO — Service Level Objective

An SLO is a **target value or range** for an SLI over a specified window.

```
SLO = SLI ≥ target  over  rolling_window
```

Example: *"99.9% of requests must complete within 200 ms over any 30-day window"*

SLOs are **internal** commitments — they should be stricter than SLAs to give you a safety margin.

```python
@dataclass
class SLO:
    name: str
    sli: SLIWindow
    target: float          # e.g. 0.999 for 99.9%
    window_days: int = 30

    @property
    def is_met(self) -> bool:
        return self.sli.value >= self.target

    @property
    def error_budget_remaining(self) -> float:
        """Fraction of allowable bad events remaining."""
        allowed_error_rate = 1.0 - self.target
        actual_error_rate  = 1.0 - self.sli.value
        if allowed_error_rate == 0:
            return 0.0
        consumed = actual_error_rate / allowed_error_rate
        return max(0.0, 1.0 - consumed)

    def report(self) -> None:
        status = "MET" if self.is_met else "BREACHED"
        print(f"SLO  : {self.name}")
        print(f"Target: {self.target*100:.2f}%  |  Actual: {self.sli.percentage:.3f}%")
        print(f"Status: {status}")
        print(f"Error budget remaining: {self.error_budget_remaining*100:.1f}%")


slo = SLO(
    name="api_latency_slo",
    sli=latency_sli,
    target=0.999,
    window_days=30,
)
slo.report()
```

## 3. SLA — Service Level Agreement

An SLA is a **contractual promise** to customers, usually with financial penalties for breach.

| | SLI | SLO | SLA |
|--|-----|-----|-----|
| What | Measurement | Target | Contract |
| Who sets it | Engineering | Engineering | Business + Legal |
| Consequence of breach | Data point | Alert, freeze features | Credits, penalties |
| Typical value | 99.95% | 99.9% | 99.5% |

**Best practice**: SLO should be tighter than SLA by at least one 9 (e.g., SLA=99.5%, SLO=99.9%).

## 4. Error Budgets

The **error budget** is the amount of unreliability you are allowed before breaching your SLO.

```
error_budget = (1 - SLO_target) × window_duration
```

For a 99.9% monthly SLO:
- 30 days × 24 hours × 60 minutes = 43,200 minutes
- Error budget = 0.1% × 43,200 = **43.2 minutes** of allowed downtime per month

When budget is full → ship features aggressively.  
When budget is exhausted → freeze releases, focus on reliability.

```python
def error_budget_minutes(slo_target: float, window_days: int = 30) -> float:
    total_minutes = window_days * 24 * 60
    return (1.0 - slo_target) * total_minutes


nines_table = [
    ("99%",    0.99),
    ("99.5%",  0.995),
    ("99.9%",  0.999),
    ("99.95%", 0.9995),
    ("99.99%", 0.9999),
]

print(f"{'SLO':<10} {'Monthly budget (min)':>22} {'Daily budget (min)':>20}")
print("-" * 55)
for label, target in nines_table:
    monthly = error_budget_minutes(target, 30)
    daily   = error_budget_minutes(target, 1)
    print(f"{label:<10} {monthly:>22.1f} {daily:>20.2f}")
```

## 5. Toil

**Toil** is work that is:
- Manual
- Repetitive
- Automatable
- Tactical (no enduring value)
- Scales with service growth

Google's SRE book suggests spending **<50% of time on toil**. Anything more erodes the team's ability to improve the system.

Examples of toil: manually restarting services, rotating credentials by hand, copying data between systems, answering tickets that could be self-served.

```python
# Toil tracker: classify work items
from dataclasses import dataclass
from typing import List


@dataclass
class WorkItem:
    description: str
    hours: float
    is_toil: bool


def toil_percentage(items: List[WorkItem]) -> float:
    total = sum(w.hours for w in items)
    toil  = sum(w.hours for w in items if w.is_toil)
    return (toil / total * 100) if total > 0 else 0.0


sprint_work = [
    WorkItem("Design new rate-limiting feature",  8.0, is_toil=False),
    WorkItem("Manually restart crashed pods",      2.0, is_toil=True),
    WorkItem("Write postmortem for DB incident",   3.0, is_toil=False),
    WorkItem("Rotate API keys for 3 services",     1.5, is_toil=True),
    WorkItem("Answer oncall tickets",              4.0, is_toil=True),
    WorkItem("Implement SLO dashboard automation", 6.0, is_toil=False),
]

pct = toil_percentage(sprint_work)
print(f"Toil percentage this sprint: {pct:.1f}%")
if pct > 50:
    print("WARNING: toil exceeds 50% — prioritise automation!")
else:
    print("Toil is within acceptable range.")
```

## 6. SLO Calculator — putting it all together

```python
class SLOCalculator:
    """Full SLO calculator: tracks events, computes budget burn rate."""

    def __init__(self, name: str, target: float, window_days: int = 30):
        self.name = name
        self.target = target
        self.window_days = window_days
        self._good = 0
        self._total = 0

    def record_event(self, good: bool) -> None:
        self._total += 1
        if good:
            self._good += 1

    @property
    def sli(self) -> float:
        return self._good / self._total if self._total else 1.0

    @property
    def total_budget_events(self) -> int:
        """Allowed bad events given current total volume."""
        return int(self._total * (1 - self.target))

    @property
    def bad_events(self) -> int:
        return self._total - self._good

    @property
    def budget_consumed_pct(self) -> float:
        if self.total_budget_events == 0:
            return 100.0
        return min(100.0, self.bad_events / self.total_budget_events * 100)

    def summary(self) -> None:
        print(f"\n{'='*50}")
        print(f"SLO Calculator: {self.name}")
        print(f"  Target SLO   : {self.target*100:.3f}%")
        print(f"  Current SLI  : {self.sli*100:.3f}%")
        print(f"  SLO {'MET' if self.sli >= self.target else 'BREACHED':8}")
        print(f"  Bad events   : {self.bad_events} / {self.total_budget_events} allowed")
        print(f"  Budget used  : {self.budget_consumed_pct:.1f}%")


# Simulate 30 days of traffic
calc = SLOCalculator("checkout_service", target=0.999)
random.seed(7)
for _ in range(500_000):
    calc.record_event(good=random.random() < 0.9991)  # slightly above target

calc.summary()
```

## Key Takeaways

- **SLI** measures the thing; **SLO** sets the target; **SLA** is the external promise
- Error budgets translate abstract reliability targets into actionable spending decisions
- Toil must be kept below 50% — SRE teams exist to *reduce* operational work, not to absorb it indefinitely
- Always set SLO tighter than SLA to maintain a safety buffer

## 🧠 Quick Quiz

**Q1.** The four Golden Signals (Google SRE) are:
- A) CPU, Memory, Disk, Network
- B) **Latency, Traffic, Errors, Saturation** ✓
- C) Availability, Reliability, Scalability, Maintainability
- D) P50, P95, P99, P999

**Q2.** An SLO (Service Level Objective):
- A) Is a legal commitment to customers
- B) **Is a target for a measurable user-facing metric (e.g. 99.9% availability)** ✓
- C) Equals the SLA
- D) Measures internal system metrics only

**Q3.** The error budget is consumed when:
- A) Development adds new features
- B) **The service falls below its SLO target** ✓
- C) Deployments are made
- D) Incidents are created

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
