Kubernetes autoscaling operates at two levels: **HPA (Horizontal Pod Autoscaler)** adds or removes pod replicas based on observed metrics, while **VPA (Vertical Pod Autoscaler)** recommends right-sized resource requests based on historical usage. Together they form a complete resource optimization loop.

**Learning objectives**
- Implement the HPA control loop algorithm from the Kubernetes source
- Model stabilization windows that prevent oscillation
- Build a VPA percentile recommender from simulated usage data
- Simulate multi-metric HPA with custom application metrics

## Core Concepts

| Concept | Definition | Key Detail |
|---|---|---|
| HPA | Adjusts replica count based on metrics | Target: CPU%, memory, custom metrics |
| Scaling ratio | `currentMetric / desiredMetric` | > 1 → scale up; < 1 → scale down |
| Stabilization window | Delay before scaling to reduce thrash | Default: 0s up, 300s down |
| Custom metrics | Non-CPU metrics via metrics-server adapter | Queue depth, RPS, GPU utilization |
| VPA | Right-sizes resource requests/limits | Uses p95 of observed usage |
| Safety margin | VPA adds buffer above observed p95 | Default: 1.15× |
| KEDA | Event-driven HPA using Kubernetes events | Scales to zero on empty queue |
| Resource pressure | Node-level constraint triggering eviction | QoS: Guaranteed > Burstable > BestEffort |

```python
import math
from dataclasses import dataclass, field
from typing import List, Dict, Optional, Tuple
from collections import deque

# ---------------------------------------------------------------------------
# HPA Controller with stabilization windows
# ---------------------------------------------------------------------------

@dataclass
class HPASpec:
    name: str
    target_deployment: str
    min_replicas: int = 2
    max_replicas: int = 10
    target_cpu_utilization: float = 70.0  # percentage
    scale_up_stabilization_s: int = 0     # seconds
    scale_down_stabilization_s: int = 300 # seconds — prevents thrash
    tolerance: float = 0.1                # ignore changes < 10%


@dataclass
class HPAController:
    spec: HPASpec
    current_replicas: int
    _up_window: deque = field(default_factory=deque)    # (timestamp, desired)
    _down_window: deque = field(default_factory=deque)  # (timestamp, desired)
    _history: List[Dict] = field(default_factory=list)

    def _compute_raw_desired(self, cpu_pct: float) -> int:
        ratio = cpu_pct / self.spec.target_cpu_utilization
        # Skip scaling if within tolerance band
        if abs(ratio - 1.0) < self.spec.tolerance:
            return self.current_replicas
        desired = math.ceil(self.current_replicas * ratio)
        return max(self.spec.min_replicas, min(self.spec.max_replicas, desired))

    def _apply_stabilization(self, raw_desired: int, now_s: int) -> int:
        """Return the most conservative value within each stabilization window."""
        if raw_desired > self.current_replicas:
            # Scale UP: keep max desired within window
            self._up_window.append((now_s, raw_desired))
            cutoff = now_s - self.spec.scale_up_stabilization_s
            while self._up_window and self._up_window[0][0] < cutoff:
                self._up_window.popleft()
            return max(d for _, d in self._up_window)
        elif raw_desired < self.current_replicas:
            # Scale DOWN: keep min desired within window (most conservative = largest)
            self._down_window.append((now_s, raw_desired))
            cutoff = now_s - self.spec.scale_down_stabilization_s
            while self._down_window and self._down_window[0][0] < cutoff:
                self._down_window.popleft()
            # Most conservative = maximum value in window (least reduction)
            return max(d for _, d in self._down_window)
        return raw_desired

    def reconcile(self, cpu_metrics: List[float], now_s: int) -> Tuple[int, str]:
        avg_cpu = sum(cpu_metrics) / len(cpu_metrics) if cpu_metrics else 0.0
        raw = self._compute_raw_desired(avg_cpu)
        stabilized = self._apply_stabilization(raw, now_s)

        action = 'NOOP'
        if stabilized > self.current_replicas:
            action = f'SCALE_UP   {self.current_replicas}→{stabilized}'
            self.current_replicas = stabilized
        elif stabilized < self.current_replicas:
            action = f'SCALE_DOWN {self.current_replicas}→{stabilized}'
            self.current_replicas = stabilized

        self._history.append({
            't': now_s, 'cpu': avg_cpu, 'raw': raw,
            'stabilized': stabilized, 'replicas': self.current_replicas,
            'action': action,
        })
        return self.current_replicas, action


# Simulate a traffic spike and recovery
hpa = HPAController(
    spec=HPASpec('inference-hpa', 'inference-api', min_replicas=2, max_replicas=12,
                 target_cpu_utilization=70, scale_down_stabilization_s=180),
    current_replicas=2,
)

# (time_seconds, cpu_samples_per_pod, description)
scenario: List[Tuple[int, List[float], str]] = [
    (0,   [35.0, 40.0],                       'baseline'),
    (30,  [60.0, 65.0],                       'moderate load'),
    (60,  [88.0, 91.0],                       'spike begins'),
    (90,  [95.0, 92.0, 88.0, 90.0],           'spike (scaled)'),
    (120, [85.0, 82.0, 88.0, 84.0, 86.0],     'peak'),
    (150, [75.0, 72.0, 78.0, 74.0, 76.0, 73.0], 'high but stable'),
    (180, [60.0, 58.0, 62.0, 55.0, 57.0, 61.0], 'recovering'),
    (240, [40.0, 38.0, 42.0, 35.0],           'load dropping'),
    (300, [28.0, 30.0, 32.0],                  'near-baseline'),
    (360, [20.0, 22.0],                         'quiet'),
    (420, [25.0, 23.0],                         'still quiet'),
    (480, [22.0, 24.0],                         'stabilized quiet (window expires)'),
]

print(f'HPA: target={hpa.spec.target_cpu_utilization}% '
      f'min={hpa.spec.min_replicas} max={hpa.spec.max_replicas} '
      f'down_window={hpa.spec.scale_down_stabilization_s}s')
print(f'{"T":>5}s  {"Avg CPU":>9}  {"Replicas":>9}  Action')
print('-' * 65)
for t, cpus, desc in scenario:
    replicas, action = hpa.reconcile(cpus, t)
    avg = sum(cpus) / len(cpus)
    print(f'  {t:4}s  {avg:7.1f}%   {replicas:4} pods  {action}  ({desc})')
```

## Multi-Metric HPA and Custom Metrics

```python
import math
from dataclasses import dataclass, field
from typing import List, Dict, Tuple

@dataclass
class MetricTarget:
    name: str
    target_value: float
    weight: float = 1.0  # for weighted combination

@dataclass
class MultiMetricHPA:
    """
    Multi-metric HPA: computes desired replicas per metric,
    takes the MAX across all metrics (most conservative scale-up).
    This matches Kubernetes' actual behavior.
    """
    min_replicas: int
    max_replicas: int
    metrics: List[MetricTarget]
    current_replicas: int

    def compute_desired(self, observations: Dict[str, float]) -> Tuple[int, Dict[str, int]]:
        per_metric: Dict[str, int] = {}
        for m in self.metrics:
            if m.name not in observations:
                continue
            current_val = observations[m.name]
            ratio = current_val / m.target_value
            desired = math.ceil(self.current_replicas * ratio)
            desired = max(self.min_replicas, min(self.max_replicas, desired))
            per_metric[m.name] = desired
        # Take maximum across all metrics (most conservative)
        final = max(per_metric.values()) if per_metric else self.current_replicas
        self.current_replicas = final
        return final, per_metric


# Simulate ML inference service with CPU + queue depth metrics
multi_hpa = MultiMetricHPA(
    min_replicas=2,
    max_replicas=15,
    metrics=[
        MetricTarget('cpu_percent', target_value=70.0),
        MetricTarget('requests_per_pod', target_value=100.0),  # custom metric
        MetricTarget('inference_queue_depth', target_value=50.0),  # KEDA-style
    ],
    current_replicas=3,
)

# Time series: (cpu%, rps_per_pod, queue_depth)
time_series = [
    (45.0, 80.0,  20.0,  'Normal'),
    (60.0, 110.0, 40.0,  'RPS slightly high'),
    (72.0, 145.0, 90.0,  'Queue backup'),
    (85.0, 160.0, 150.0, 'All metrics high'),
    (55.0, 95.0,  60.0,  'CPU down, queue lingering'),
    (40.0, 70.0,  10.0,  'Recovery'),
]

print(f'Multi-metric HPA: min={multi_hpa.min_replicas} max={multi_hpa.max_replicas}')
print(f'Targets: cpu=70% | rps_per_pod=100 | queue_depth=50')
print()
print(f'{"CPU%":>7}  {"RPS/pod":>8}  {"Queue":>7}  {"Replicas":>9}  {"Metric Breakdown"}')
print('-' * 80)
for cpu, rps, queue, desc in time_series:
    obs = {'cpu_percent': cpu, 'requests_per_pod': rps, 'inference_queue_depth': queue}
    final, per_metric = multi_hpa.compute_desired(obs)
    breakdown = ' | '.join(f'{k}→{v}r' for k, v in per_metric.items())
    print(f'  {cpu:5.1f}%  {rps:8.1f}  {queue:7.1f}  {final:4} pods  [{breakdown}]  {desc}')
```

## VPA Percentile Recommender

```python
import random
import statistics
from dataclasses import dataclass
from typing import List

@dataclass
class VPARecommendation:
    container: str
    cpu_request_m: int   # millicores
    memory_request_mi: int  # MiB
    cpu_limit_m: int
    memory_limit_mi: int
    confidence: str  # Low / Medium / High

    def show(self) -> None:
        print(f'  VPA for container: {self.container}  [{self.confidence} confidence]')
        print(f'    cpu:    request={self.cpu_request_m}m      limit={self.cpu_limit_m}m')
        print(f'    memory: request={self.memory_request_mi}Mi    limit={self.memory_limit_mi}Mi')


def percentile(data: List[float], p: float) -> float:
    if not data:
        return 0.0
    s = sorted(data)
    idx = max(0, int(len(s) * p / 100) - 1)
    return s[idx]


def vpa_recommend(
    container: str,
    cpu_samples_m: List[float],      # millicores
    mem_samples_mi: List[float],     # MiB
    safety_margin: float = 1.15,
    limit_multiplier: float = 2.0,
) -> VPARecommendation:
    n = len(cpu_samples_m)
    confidence = 'Low' if n < 100 else ('Medium' if n < 500 else 'High')

    cpu_p95  = percentile(cpu_samples_m, 95)
    mem_p95  = percentile(mem_samples_mi, 95)

    cpu_req  = int(cpu_p95 * safety_margin)
    mem_req  = int(mem_p95 * safety_margin)
    cpu_lim  = int(cpu_req * limit_multiplier)
    mem_lim  = int(mem_req * limit_multiplier)

    return VPARecommendation(container, cpu_req, mem_req, cpu_lim, mem_lim, confidence)


# Simulate 7 days of per-pod CPU and memory samples (one sample per minute)
rng = random.Random(42)
n_samples = 10_080  # 7 days * 24 hours * 60 minutes

# Bimodal: mostly idle with occasional inference spikes
cpu_samples = []
mem_samples = []
for _ in range(n_samples):
    if rng.random() < 0.15:  # 15% chance of active inference
        cpu_samples.append(max(10.0, rng.gauss(600, 120)))  # spike: ~600m
        mem_samples.append(max(100.0, rng.gauss(750, 100)))
    else:
        cpu_samples.append(max(5.0, rng.gauss(80, 30)))     # idle: ~80m
        mem_samples.append(max(100.0, rng.gauss(320, 50)))

rec = vpa_recommend('inference-api', cpu_samples, mem_samples)

print('Current (over-provisioned defaults):')
print('    cpu:    request=100m       limit=2000m')
print('    memory: request=128Mi      limit=2048Mi')
print()
print('VPA recommendation after 7-day observation:')
rec.show()

# Show percentile distribution
print()
print('CPU distribution (millicores):')
for p in [50, 75, 90, 95, 99]:
    print(f'    p{p:2}: {percentile(cpu_samples, p):.0f}m')

print()
print('Memory distribution (MiB):')
for p in [50, 75, 90, 95, 99]:
    print(f'    p{p:2}: {percentile(mem_samples, p):.0f}Mi')
```

## Quick Quiz

**Q1**: What does the HPA stabilization window prevent?
> **A**: Scale thrashing — rapidly cycling replicas up and down when metrics fluctuate near the target. The scale-down window (default 5 minutes) holds the replica count high until metrics have consistently dropped below target, preventing unnecessary churn.

**Q2**: When is VPA preferred over HPA?
> **A**: VPA is appropriate for workloads that cannot scale horizontally: single-instance databases, stateful services with sticky sessions, or jobs running on expensive GPU nodes where adding more replicas would exceed quota. VPA also complements HPA for stateless services — VPA right-sizes requests (improving scheduler bin-packing) while HPA controls replica count.

**Q3**: Why does multi-metric HPA take the MAX desired replicas across all metrics?
> **A**: Taking the max is conservative for scale-up: if any metric is overloaded, the service needs more capacity. It would be wrong to average — a service where CPU is fine but the request queue is exploding still needs more pods.

## Key Takeaways

- HPA formula: `desired = ceil(current * currentMetric/targetMetric)` clamped to `[min, max]`
- Stabilization windows are sliding window max/min operations that dampen oscillation
- Multi-metric HPA takes the maximum desired replicas across all metrics
- VPA uses p95 of observed usage + safety margin to set right-sized requests
- CPU is compressible (throttled at limit); memory is incompressible (OOMKilled at limit)

## Going Deeper

1. **Exercise**: Add a `tolerance` band to `HPAController` — don't scale if `|ratio - 1.0| < 0.1`. Plot replica count vs time with and without the tolerance band.
2. **Exercise**: Implement `KEDAScaler` that scales to zero when queue depth = 0 and scales from zero when a first item arrives. What's the cold-start problem and how do you mitigate it?
3. **Exercise**: Model a pathological feedback loop — HPA scales down → CPU spikes → HPA scales up → CPU drops → repeat. What stabilization window length stops the oscillation?
4. **Project**: Combine VPA recommendations with HPA: use VPA-suggested requests so HPA's CPU% calculations are meaningful. Show how wrong resource requests make HPA fire incorrectly.
5. **Research**: Read the KEDA project docs. How does a `ScaledObject` with `cooldownPeriod` + `pollingInterval` differ from a native HPA with `stabilizationWindowSeconds`?
