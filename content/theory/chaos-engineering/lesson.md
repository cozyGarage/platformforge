Topics: chaos principles · fault injection patterns · steady-state hypothesis · chaos runner

## 1. What is Chaos Engineering?

> *"Chaos Engineering is the discipline of experimenting on a system in order to build confidence in the system's capability to withstand turbulent conditions in production."*  
> — Principles of Chaos Engineering (principlesofchaos.org)

Key idea: **proactively inject failures** in controlled conditions to find weaknesses before they become incidents.

Chaos is not random destruction — it is **scientific experimentation** with:
1. A hypothesis about system behaviour
2. A controlled blast radius
3. Automated abort conditions
4. Observable outcomes

## 2. The Five Principles

1. **Define steady state** — quantify "normal" system behaviour (SLIs)
2. **Hypothesise** — predict the system will remain in steady state during the experiment
3. **Introduce real-world events** — latency, failures, resource exhaustion
4. **Disprove the hypothesis** — if steady state breaks, you found a weakness
5. **Minimise blast radius** — start with small scope (canary, shadow traffic, test env)

Common fault types:
- **Network**: latency injection, packet loss, DNS failures
- **Resource**: CPU spike, memory pressure, disk full
- **Dependency**: kill upstream service, inject 500 responses
- **Process**: kill random instances, OOM kill

```python
import time
import random
from dataclasses import dataclass, field
from typing import Callable, Dict, List, Optional, Tuple
from enum import Enum


class FaultType(Enum):
    LATENCY          = "latency"
    ERROR_INJECTION  = "error_injection"
    RESOURCE_LIMIT   = "resource_limit"
    DEPENDENCY_KILL  = "dependency_kill"


@dataclass
class FaultConfig:
    fault_type:   FaultType
    intensity:    float     # 0.0 – 1.0 (e.g. 0.1 = inject into 10% of calls)
    params:       Dict      = field(default_factory=dict)


class FaultInjector:
    """Wraps a callable and injects faults according to config."""

    def __init__(self, config: FaultConfig, seed: int = 0):
        self.config = config
        self._rng   = random.Random(seed)

    def call(self, fn: Callable, *args, **kwargs):
        if self._rng.random() > self.config.intensity:
            return fn(*args, **kwargs)   # pass-through

        ft = self.config.fault_type
        if ft == FaultType.LATENCY:
            delay = self.config.params.get("delay_ms", 500) / 1000
            time.sleep(delay)
            return fn(*args, **kwargs)
        elif ft == FaultType.ERROR_INJECTION:
            raise RuntimeError(self.config.params.get("message", "injected error"))
        elif ft == FaultType.DEPENDENCY_KILL:
            raise ConnectionError("dependency unavailable (injected)")
        else:
            return fn(*args, **kwargs)


# Demonstrate fault injection
def checkout_handler(order_id: str) -> Dict:
    return {"status": "ok", "order_id": order_id}


injector = FaultInjector(
    FaultConfig(FaultType.ERROR_INJECTION, intensity=0.3,
                params={"message": "payment gateway timeout"}),
    seed=42,
)

ok = errors = 0
for i in range(20):
    try:
        result = injector.call(checkout_handler, f"order-{i}")
        ok += 1
    except RuntimeError as e:
        errors += 1

print(f"20 calls — ok={ok}, errors={errors} (injected ~30%)")
```

## 3. Steady-State Hypothesis

Before running a chaos experiment, define measurable steady state:

```
Hypothesis: "During the experiment, error_rate < 1% AND p99_latency < 300ms"
```

If the hypothesis is violated, the experiment reveals a resilience gap.

```python
@dataclass
class SteadyStateProbe:
    """Checks system metrics against steady-state bounds."""
    name:      str
    check:     Callable[[Dict], bool]  # returns True = healthy
    metric_key: str


@dataclass
class ChaosExperiment:
    name:         str
    hypothesis:   str
    fault:        FaultConfig
    probes:       List[SteadyStateProbe]

    def run(self, service_fn: Callable, n_calls: int = 100) -> None:
        print(f"\nChaos Experiment: {self.name}")
        print(f"Hypothesis: {self.hypothesis}")
        print("-" * 55)

        injector = FaultInjector(self.fault, seed=77)
        results = {"ok": 0, "errors": 0, "latencies_ms": []}

        for i in range(n_calls):
            t0 = time.monotonic()
            try:
                injector.call(service_fn, f"req-{i}")
                results["ok"] += 1
            except Exception:
                results["errors"] += 1
            results["latencies_ms"].append((time.monotonic() - t0) * 1000)

        lats = sorted(results["latencies_ms"])
        p99  = lats[int(len(lats) * 0.99)]
        err_rate = results["errors"] / n_calls

        metrics = {"error_rate": err_rate, "p99_latency_ms": p99}
        print(f"  error_rate   : {err_rate*100:.1f}%")
        print(f"  p99_latency  : {p99:.1f} ms")

        all_healthy = all(p.check(metrics) for p in self.probes)
        verdict = "HYPOTHESIS CONFIRMED" if all_healthy else "HYPOTHESIS DISPROVED — weakness found!"
        print(f"  Result: {verdict}")


experiment = ChaosExperiment(
    name="Payment gateway latency injection",
    hypothesis="System maintains <1% errors and <300ms p99 under 20% latency injection",
    fault=FaultConfig(FaultType.LATENCY, intensity=0.2, params={"delay_ms": 400}),
    probes=[
        SteadyStateProbe("error_rate", lambda m: m["error_rate"] < 0.01, "error_rate"),
        SteadyStateProbe("p99_latency", lambda m: m["p99_latency_ms"] < 300, "p99_latency_ms"),
    ],
)

experiment.run(checkout_handler, n_calls=50)
```

## 4. Blast Radius Control

Always limit blast radius:

| Technique | How |
|-----------|-----|
| **Shadow traffic** | Replay production traffic to a chaos target (no user impact) |
| **Canary targeting** | Apply fault only to 1-5% of instances |
| **Time-boxed** | Run for 5 minutes, auto-stop |
| **Abort conditions** | If error_rate > 5%, terminate experiment immediately |
| **Non-peak hours** | Run experiments when traffic volume is lowest |

```python
class SafeChaosRunner:
    """Runs chaos experiments with automatic abort on breach."""

    def __init__(self, max_error_rate: float = 0.05, max_duration_s: float = 30):
        self.max_error_rate  = max_error_rate
        self.max_duration_s  = max_duration_s

    def run(self, experiment: ChaosExperiment, service_fn: Callable,
            n_calls: int = 200) -> None:
        print(f"\nSafe Chaos Run: {experiment.name}")
        injector   = FaultInjector(experiment.fault, seed=100)
        errors = ok = 0
        aborted = False
        t_start = time.monotonic()

        for i in range(n_calls):
            if time.monotonic() - t_start > self.max_duration_s:
                print(f"  [ABORT] Time limit {self.max_duration_s}s reached after {i} calls")
                aborted = True
                break

            total = ok + errors
            if total > 10 and errors / total > self.max_error_rate:
                print(f"  [ABORT] Error rate {errors/total*100:.1f}% exceeded {self.max_error_rate*100:.0f}% threshold")
                aborted = True
                break

            try:
                injector.call(service_fn, f"req-{i}")
                ok += 1
            except Exception:
                errors += 1

        total = ok + errors
        print(f"  Completed: {total} calls, {errors} errors ({errors/total*100:.1f}%)")
        print(f"  Aborted early: {aborted}")


safe_runner = SafeChaosRunner(max_error_rate=0.25, max_duration_s=60)

safe_exp = ChaosExperiment(
    name="Error injection at 40% — expect abort",
    hypothesis="System healthy under 40% error injection",
    fault=FaultConfig(FaultType.ERROR_INJECTION, intensity=0.40),
    probes=[],
)
safe_runner.run(safe_exp, checkout_handler, n_calls=100)
```

## Key Takeaways

- Chaos engineering is **scientific** — hypothesis, experiment, measurement, conclusion
- Always define **steady-state** before the experiment — you need a baseline to compare against
- Control **blast radius**: canary, time-box, abort conditions
- Inject real-world failure modes: latency, errors, dependency kills, resource pressure
- Finding weaknesses in staging is far cheaper than discovering them during incidents

## 🧠 Quick Quiz

**Q1.** Chaos engineering's core principle is:
- A) Breaking things randomly in production
- B) **Proactively injecting controlled failures to discover weaknesses before they cause incidents** ✓
- C) Disabling monitoring during tests
- D) Only applicable to large organisations

**Q2.** A "Game Day" in SRE is:
- A) A team social event
- B) **A planned exercise simulating a major failure to test response procedures** ✓
- C) A sprint retrospective
- D) A scheduled maintenance window

**Q3.** The blast radius in chaos experiments should be:
- A) As large as possible for realism
- B) **Minimised and controlled — start small, expand incrementally** ✓
- C) Unlimited in staging environments
- D) Determined by random chance

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
