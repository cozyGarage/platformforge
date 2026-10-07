Topics: alert fatigue · symptom vs cause alerts · runbooks · incident severity · alert evaluator

## 1. Alert Fatigue

**Alert fatigue** occurs when on-call engineers receive so many alerts that they begin ignoring them.  
It is one of the most dangerous failure modes in an SRE organisation.

Causes:
- Alerts that fire but require no action (informational noise)
- Duplicate alerts for the same root cause
- Flapping alerts that oscillate around a threshold
- Overly tight thresholds (too sensitive)

The fix: **every alert that pages someone must be actionable and urgent**.

## 2. Symptom vs Cause Alerts

**Cause-based** alerts fire on individual infrastructure signals:
- CPU > 80%
- Disk > 90%
- Process restarted

**Symptom-based** alerts fire on user-visible impact:
- Error rate > 1% (SLO burn rate)
- p99 latency > 500 ms
- Checkout conversion dropped

**Best practice**: page on symptoms, use cause-based alerts for tickets/dashboards only.  
The question to ask: *"Is a user harmed right now?"*

```python
from dataclasses import dataclass, field
from typing import Callable, Dict, List, Optional, Any
from enum import Enum
import datetime


class Severity(Enum):
    PAGE   = "PAGE"    # wake someone up immediately
    TICKET = "TICKET"  # create a ticket, handle next business day
    INFO   = "INFO"    # log only, no action required


@dataclass
class AlertRule:
    name:        str
    description: str
    severity:    Severity
    is_symptom:  bool
    condition:   Callable[[Dict[str, float]], bool]
    runbook_url: str = ""


@dataclass
class FiredAlert:
    rule:      AlertRule
    fired_at:  str
    context:   Dict[str, float]

    def __str__(self) -> str:
        return (
            f"[{self.rule.severity.value}] {self.rule.name} @ {self.fired_at}\n"
            f"  {self.rule.description}\n"
            f"  context: {self.context}\n"
            f"  runbook: {self.rule.runbook_url or 'N/A'}"
        )


class AlertEvaluator:
    """Evaluates a set of alert rules against current metric values."""

    def __init__(self):
        self._rules: List[AlertRule] = []

    def register(self, rule: AlertRule) -> None:
        self._rules.append(rule)

    def evaluate(self, metrics: Dict[str, float]) -> List[FiredAlert]:
        fired = []
        now = datetime.datetime.utcnow().isoformat() + "Z"
        for rule in self._rules:
            if rule.condition(metrics):
                fired.append(FiredAlert(rule=rule, fired_at=now, context=metrics))
        return fired


# Define rules
evaluator = AlertEvaluator()

evaluator.register(AlertRule(
    name="high_error_rate",
    description="Error rate exceeds SLO burn threshold",
    severity=Severity.PAGE,
    is_symptom=True,
    condition=lambda m: m.get("error_rate", 0) > 0.01,
    runbook_url="https://wiki/runbooks/high-error-rate",
))

evaluator.register(AlertRule(
    name="high_p99_latency",
    description="p99 latency > 500 ms",
    severity=Severity.PAGE,
    is_symptom=True,
    condition=lambda m: m.get("p99_latency_ms", 0) > 500,
    runbook_url="https://wiki/runbooks/high-latency",
))

evaluator.register(AlertRule(
    name="high_cpu",
    description="CPU utilisation > 80%",
    severity=Severity.TICKET,
    is_symptom=False,
    condition=lambda m: m.get("cpu_pct", 0) > 80,
))

# Evaluate against current state
current_metrics = {"error_rate": 0.025, "p99_latency_ms": 320, "cpu_pct": 85}
alerts = evaluator.evaluate(current_metrics)

print(f"Fired {len(alerts)} alert(s):")
for a in alerts:
    print()
    print(a)
```

## 3. Incident Severity Classification

| Severity | Impact | Response time | Example |
|----------|--------|---------------|---------|
| **SEV-1** | Total outage, all users affected | Immediate | Payment system down |
| **SEV-2** | Partial outage, major feature broken | < 15 min | Search unavailable |
| **SEV-3** | Degraded performance, workaround exists | < 1 hour | Slow image uploads |
| **SEV-4** | Minor issue, no user impact | Next business day | Flapping health check |

Severity drives **escalation policy** and **communication cadence**.

```python
def classify_severity(
    error_rate: float,
    affected_users_pct: float,
    feature_unavailable: bool,
) -> int:
    """Classify incident severity 1-4."""
    if error_rate > 0.5 or affected_users_pct > 50:
        return 1
    if feature_unavailable or error_rate > 0.1 or affected_users_pct > 10:
        return 2
    if error_rate > 0.01 or affected_users_pct > 1:
        return 3
    return 4


scenarios = [
    dict(error_rate=0.60, affected_users_pct=70, feature_unavailable=True),
    dict(error_rate=0.15, affected_users_pct=20, feature_unavailable=True),
    dict(error_rate=0.02, affected_users_pct=3,  feature_unavailable=False),
    dict(error_rate=0.001, affected_users_pct=0, feature_unavailable=False),
]

for s in scenarios:
    sev = classify_severity(**s)
    print(f"SEV-{sev}: error={s['error_rate']*100:.0f}%  users={s['affected_users_pct']}%  "
          f"unavailable={s['feature_unavailable']}")
```

## 4. Runbook Automation

A **runbook** is a documented set of steps for responding to a specific alert.  
Automated runbooks can execute diagnostic commands and even remediation steps.

Structure of a good runbook:
1. **What fired** — alert name and threshold
2. **Why it matters** — user impact
3. **Immediate checks** — what to look at first
4. **Remediation steps** — ordered, with expected outcomes
5. **Escalation path** — who to page if steps fail

```python
@dataclass
class RunbookStep:
    description: str
    action:      Callable[[], str]   # returns outcome message


class Runbook:
    def __init__(self, name: str):
        self.name  = name
        self.steps: List[RunbookStep] = []

    def add_step(self, description: str, action: Callable[[], str]) -> None:
        self.steps.append(RunbookStep(description=description, action=action))

    def execute(self) -> None:
        print(f"\nRunbook: {self.name}")
        print("=" * 50)
        for i, step in enumerate(self.steps, 1):
            print(f"Step {i}: {step.description}")
            result = step.action()
            print(f"  => {result}")


# Simulate a high-error-rate runbook
import random as _r
_r.seed(5)

rb = Runbook("high_error_rate_checkout")
rb.add_step("Check recent deployments",
            lambda: "Last deploy: checkout-v2.3.1, 12 minutes ago")
rb.add_step("Check DB connection pool",
            lambda: f"Pool utilisation: {_r.randint(60,95)}%")
rb.add_step("Check upstream payment gateway",
            lambda: "Stripe status: degraded performance (official status page)")
rb.add_step("Enable fallback payment method",
            lambda: "Fallback enabled: PayPal gateway active")

rb.execute()
```

## Key Takeaways

- **Alert fatigue kills** — every page must be actionable, urgent, and non-duplicated
- **Symptom-based alerts** (user harm) should page; cause-based alerts should create tickets
- Severity classification drives response time and communication cadence
- Runbooks should be **executable** — automated diagnostic steps save minutes during incidents
- Review alert signal/noise ratio monthly; delete or demote any alert that fires without action

## 🧠 Quick Quiz

**Q1.** Alert fatigue occurs when:
- A) Alerts are too infrequent
- B) **Too many low-quality alerts desensitise oncall responders to real incidents** ✓
- C) Alerting systems are unreliable
- D) Teams are too small

**Q2.** A good SLO (Service Level Objective) should be:
- A) Set at 100% availability
- B) **Achievable, meaningful to users, and based on error budget thinking** ✓
- C) Changed monthly
- D) The same as SLA

**Q3.** PagerDuty escalation policies ensure:
- A) The fastest engineer is always notified first
- B) **If the primary oncall doesn't respond, the alert escalates to the next person** ✓
- C) All team members receive all alerts
- D) Alerts are automatically resolved

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
