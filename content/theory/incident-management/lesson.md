Topics: incident lifecycle · blameless postmortems · five whys · timeline builder · MTTR/MTBF

## 1. Incident Lifecycle

```
DETECT → TRIAGE → MITIGATE → RESOLVE → POSTMORTEM
```

| Phase | Goal | Owner |
|-------|------|-------|
| **Detect** | Alert fires or user reports problem | Monitoring / on-call |
| **Triage** | Confirm severity, assign IC (Incident Commander) | On-call engineer |
| **Mitigate** | Reduce user impact (rollback, disable feature flag) | IC + responders |
| **Resolve** | Root cause fixed, service fully restored | IC + engineering |
| **Postmortem** | Learn and prevent recurrence | Team |

```python
import datetime
from dataclasses import dataclass, field
from typing import Dict, List, Optional
from enum import Enum


class IncidentPhase(Enum):
    DETECTING   = "detecting"
    TRIAGING    = "triaging"
    MITIGATING  = "mitigating"
    RESOLVED    = "resolved"
    POSTMORTEM  = "postmortem"


@dataclass
class TimelineEvent:
    timestamp: datetime.datetime
    actor:     str
    message:   str
    phase:     IncidentPhase


@dataclass
class Incident:
    id:          str
    title:       str
    severity:    int
    opened_at:   datetime.datetime
    resolved_at: Optional[datetime.datetime] = None
    timeline:    List[TimelineEvent] = field(default_factory=list)

    def add_event(self, actor: str, message: str, phase: IncidentPhase) -> None:
        event = TimelineEvent(
            timestamp=datetime.datetime.utcnow(),
            actor=actor,
            message=message,
            phase=phase,
        )
        self.timeline.append(event)

    def resolve(self) -> None:
        self.resolved_at = datetime.datetime.utcnow()

    @property
    def ttd_minutes(self) -> Optional[float]:
        """Time to detect (first triage event)."""
        for e in self.timeline:
            if e.phase == IncidentPhase.TRIAGING:
                return (e.timestamp - self.opened_at).total_seconds() / 60
        return None

    @property
    def ttm_minutes(self) -> Optional[float]:
        """Time to mitigate."""
        for e in self.timeline:
            if e.phase == IncidentPhase.MITIGATING:
                return (e.timestamp - self.opened_at).total_seconds() / 60
        return None


# Create a sample incident
import time

inc = Incident(
    id="INC-2024-001",
    title="Checkout error rate spike — 40% errors on payment endpoint",
    severity=1,
    opened_at=datetime.datetime.utcnow(),
)

inc.add_event("monitoring",    "alert fired: error_rate > 1%",              IncidentPhase.DETECTING)
time.sleep(0.01)
inc.add_event("alice",         "acknowledged, starting triage",             IncidentPhase.TRIAGING)
time.sleep(0.01)
inc.add_event("alice",         "identified: v2.3.1 deploy 15 min ago",      IncidentPhase.TRIAGING)
time.sleep(0.01)
inc.add_event("alice",         "rolling back to v2.3.0",                    IncidentPhase.MITIGATING)
time.sleep(0.01)
inc.add_event("monitoring",    "error rate returned to baseline",           IncidentPhase.RESOLVED)
inc.resolve()

print(f"Incident: {inc.id} — SEV-{inc.severity}")
print(f"Title   : {inc.title}")
print()
for e in inc.timeline:
    print(f"  [{e.phase.value:10}] {e.actor:<12}: {e.message}")
```

## 2. Blameless Postmortems

A **postmortem** (or post-incident review) documents what happened, why, and how to prevent recurrence.

**Blameless** means: assume people acted with the best intentions given the information available at the time. Focus on systemic fixes, not individual errors.

Anatomy of a postmortem:
1. **Summary** — 2-3 sentence description of impact
2. **Timeline** — chronological events
3. **Root cause analysis** — what actually caused the failure
4. **Contributing factors** — environmental conditions
5. **Impact** — user/revenue/reliability numbers
6. **Action items** — specific, owned, time-bounded

```python
def five_whys(initial_symptom: str, whys: List[str]) -> str:
    """Format a five-whys root cause analysis."""
    lines = [f"Symptom: {initial_symptom}", ""]
    for i, why in enumerate(whys, 1):
        lines.append(f"Why {i}: {why}")
    lines.append("")
    lines.append(f"Root cause: {whys[-1]}")
    return "\n".join(lines)


analysis = five_whys(
    initial_symptom="40% of checkout requests returned 500 errors",
    whys=[
        "The payment service was returning database connection errors",
        "The connection pool was exhausted (all 20 connections in use)",
        "A slow DB query introduced in v2.3.1 held connections for 30s",
        "The query was missing an index on orders.user_id",
        "The migration adding that index was omitted from the deploy checklist",
    ],
)
print(analysis)
```

## 3. MTTR & MTBF

**MTTR** — Mean Time To Recover: average time from incident start to resolution  
**MTBF** — Mean Time Between Failures: average time between incident resolutions

```
MTTR = total_downtime / incident_count
MTBF = total_uptime / incident_count
Availability = MTBF / (MTBF + MTTR)
```

Low MTTR → fast recovery (good runbooks, good tooling)  
High MTBF → infrequent failures (good testing, gradual rollouts)

```python
from typing import Tuple


@dataclass
class IncidentRecord:
    started_at:  datetime.datetime
    resolved_at: datetime.datetime

    @property
    def duration_minutes(self) -> float:
        return (self.resolved_at - self.started_at).total_seconds() / 60


def compute_reliability_metrics(
    incidents: List[IncidentRecord],
    observation_period_hours: float,
) -> Dict[str, float]:
    if not incidents:
        return {"mttr_min": 0, "mtbf_hours": observation_period_hours, "availability": 1.0}

    total_downtime_min = sum(i.duration_minutes for i in incidents)
    mttr = total_downtime_min / len(incidents)

    total_uptime_hours = observation_period_hours - (total_downtime_min / 60)
    mtbf = total_uptime_hours / len(incidents) if len(incidents) > 0 else total_uptime_hours

    mttr_hours = mttr / 60
    availability = mtbf / (mtbf + mttr_hours) if (mtbf + mttr_hours) > 0 else 1.0

    return {
        "incident_count":   len(incidents),
        "total_downtime_min": total_downtime_min,
        "mttr_min":         mttr,
        "mtbf_hours":       mtbf,
        "availability_pct": availability * 100,
    }


# 30-day incident log
base = datetime.datetime(2024, 1, 1)
records = [
    IncidentRecord(base + datetime.timedelta(days=3,  hours=2),
                   base + datetime.timedelta(days=3,  hours=2, minutes=45)),
    IncidentRecord(base + datetime.timedelta(days=11, hours=14),
                   base + datetime.timedelta(days=11, hours=14, minutes=12)),
    IncidentRecord(base + datetime.timedelta(days=22, hours=9),
                   base + datetime.timedelta(days=22, hours=9,  minutes=90)),
]

metrics = compute_reliability_metrics(records, observation_period_hours=30*24)
print("30-day reliability metrics")
print("-" * 35)
for k, v in metrics.items():
    print(f"  {k:<25}: {v:.2f}")
```

## Key Takeaways

- The incident lifecycle: **Detect → Triage → Mitigate → Resolve → Postmortem**
- Blameless postmortems improve system reliability; blame-driven ones destroy trust
- **Five whys** reveals systemic root causes beneath surface symptoms
- Track **MTTR** (speed of recovery) and **MTBF** (frequency of incidents) over time
- Action items from postmortems must be **specific, owned, and time-bounded** — vague items rot

## 🧠 Quick Quiz

**Q1.** A blameless post-mortem focuses on:
- A) Identifying the responsible engineer
- B) **Understanding systemic causes and improving processes to prevent recurrence** ✓
- C) Disciplinary actions
- D) SLA penalties

**Q2.** Mean Time To Detect (MTTD) measures:
- A) How long it takes to fix an incident
- B) **The average time from incident start to detection** ✓
- C) The frequency of incidents
- D) The impact of an incident

**Q3.** Runbooks in incident management provide:
- A) Architecture diagrams
- B) **Step-by-step procedures for responding to known failure scenarios** ✓
- C) Post-mortem templates
- D) SLA definitions

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
