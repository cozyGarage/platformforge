Topics: Prometheus data model · counters · gauges · histograms · RED · USE

## 1. Prometheus Data Model

Prometheus represents every time series as:

```
metric_name{label1="value1", label2="value2"} <float64_value> [<unix_timestamp_ms>]
```

**Labels** add dimensionality — you can slice and aggregate by any combination.  
High-cardinality labels (e.g. user_id) are an anti-pattern — they explode storage.

Core metric types:
| Type | Monotonic? | Use for |
|------|-----------|--------|
| Counter | Yes (only up) | Total requests, errors, bytes sent |
| Gauge | No (up/down) | CPU usage, queue depth, active connections |
| Histogram | Yes (observations) | Request latency distributions |
| Summary | Yes (observations) | Pre-computed quantiles (less flexible) |

```python
import time
import math
from dataclasses import dataclass, field
from typing import Dict, List, Tuple, Optional
from collections import defaultdict


Labels = Dict[str, str]


def _labels_key(labels: Labels) -> str:
    return ",".join(f'{k}="{v}"' for k, v in sorted(labels.items()))


class Counter:
    """Monotonically increasing counter."""

    def __init__(self, name: str, help_text: str = ""):
        self.name = name
        self.help = help_text
        self._values: Dict[str, float] = defaultdict(float)

    def inc(self, amount: float = 1.0, labels: Optional[Labels] = None) -> None:
        if amount < 0:
            raise ValueError("Counter can only be incremented")
        key = _labels_key(labels or {})
        self._values[key] += amount

    def get(self, labels: Optional[Labels] = None) -> float:
        return self._values[_labels_key(labels or {})]

    def __repr__(self) -> str:
        lines = [f"# HELP {self.name} {self.help}", f"# TYPE {self.name} counter"]
        for labels_key, val in self._values.items():
            if labels_key:
                lines.append(f"{self.name}{{{labels_key}}} {val}")
            else:
                lines.append(f"{self.name} {val}")
        return "\n".join(lines)


# Usage
http_requests = Counter("http_requests_total", "Total HTTP requests")
http_requests.inc(labels={"method": "GET",  "status": "200"})
http_requests.inc(labels={"method": "GET",  "status": "200"})
http_requests.inc(labels={"method": "POST", "status": "500"})

print(http_requests)
```

```python
class Gauge:
    """Gauge that can go up and down."""

    def __init__(self, name: str, help_text: str = ""):
        self.name = name
        self.help = help_text
        self._values: Dict[str, float] = defaultdict(float)

    def set(self, value: float, labels: Optional[Labels] = None) -> None:
        self._values[_labels_key(labels or {})] = value

    def inc(self, amount: float = 1.0, labels: Optional[Labels] = None) -> None:
        self._values[_labels_key(labels or {})] += amount

    def dec(self, amount: float = 1.0, labels: Optional[Labels] = None) -> None:
        self._values[_labels_key(labels or {})] -= amount

    def get(self, labels: Optional[Labels] = None) -> float:
        return self._values[_labels_key(labels or {})]


queue_depth = Gauge("job_queue_depth", "Current jobs waiting")
queue_depth.set(0)
queue_depth.inc(5)   # 5 jobs arrive
queue_depth.dec(2)   # 2 jobs processed
print(f"Queue depth: {queue_depth.get()}")  # 3
```

## 2. Histograms

Histograms count observations in configurable **buckets**, enabling percentile estimation.

Prometheus text format:
```
http_request_duration_seconds_bucket{le="0.1"}  24054
http_request_duration_seconds_bucket{le="0.5"}  33444
http_request_duration_seconds_bucket{le="1.0"}  100392
http_request_duration_seconds_bucket{le="+Inf"} 144320
http_request_duration_seconds_sum               53423
http_request_duration_seconds_count             144320
```

Use `histogram_quantile(0.99, rate(...))` in PromQL to compute p99.

```python
import random


class Histogram:
    """Prometheus-style histogram with configurable buckets."""

    DEFAULT_BUCKETS = (0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)

    def __init__(self, name: str, buckets: Tuple[float, ...] = DEFAULT_BUCKETS):
        self.name = name
        self.buckets = sorted(buckets)
        self._counts = [0] * len(self.buckets)
        self._sum = 0.0
        self._count = 0

    def observe(self, value: float) -> None:
        self._sum += value
        self._count += 1
        for i, upper in enumerate(self.buckets):
            if value <= upper:
                self._counts[i] += 1

    def percentile_estimate(self, p: float) -> float:
        """Linear interpolation estimate of given percentile (0-1)."""
        target = self._count * p
        cumulative = 0
        prev_upper = 0.0
        for upper, cnt in zip(self.buckets, self._counts):
            cumulative += cnt
            if cumulative >= target:
                return upper
            prev_upper = upper
        return self.buckets[-1]

    def report(self) -> None:
        print(f"Histogram: {self.name}")
        print(f"  count={self._count}, sum={self._sum:.3f}, mean={self._sum/self._count:.3f}")
        for p in (0.50, 0.90, 0.95, 0.99):
            est = self.percentile_estimate(p)
            print(f"  ~p{int(p*100):02d}: {est:.4f}")


# Simulate request latencies
random.seed(99)
hist = Histogram("http_request_duration_seconds")
for _ in range(10_000):
    latency = abs(random.gauss(0.08, 0.04))  # mostly fast, some outliers
    hist.observe(latency)

hist.report()
```

## 3. Metrics Registry

A registry manages all metrics and produces a scrape endpoint (like `/metrics`).

```python
class MetricsRegistry:
    """Central registry — acts as Prometheus /metrics endpoint."""

    def __init__(self):
        self._counters:   Dict[str, Counter]   = {}
        self._gauges:     Dict[str, Gauge]     = {}
        self._histograms: Dict[str, Histogram] = {}

    def register_counter(self, name: str, help_text: str = "") -> Counter:
        c = Counter(name, help_text)
        self._counters[name] = c
        return c

    def register_gauge(self, name: str, help_text: str = "") -> Gauge:
        g = Gauge(name, help_text)
        self._gauges[name] = g
        return g

    def register_histogram(self, name: str) -> Histogram:
        h = Histogram(name)
        self._histograms[name] = h
        return h

    def scrape(self) -> str:
        """Simulate /metrics text output."""
        lines = []
        for name, c in self._counters.items():
            lines.append(str(c))
        for name, g in self._gauges.items():
            lines.append(f"# TYPE {name} gauge")
            for k, v in g._values.items():
                if k:
                    lines.append(f"{name}{{{k}}} {v}")
                else:
                    lines.append(f"{name} {v}")
        return "\n".join(lines)


reg = MetricsRegistry()
reqs  = reg.register_counter("http_requests_total", "HTTP requests")
errs  = reg.register_counter("http_errors_total",   "HTTP 5xx errors")
conns = reg.register_gauge("active_connections",     "Open connections")

reqs.inc(100);  errs.inc(3);  conns.set(42)

print(reg.scrape())
```

## 4. RED Method

Coined by Tom Wilkie — for **every service** monitor:

| Letter | Metric | PromQL example |
|--------|--------|---------------|
| **R**ate | Requests/sec | `rate(http_requests_total[5m])` |
| **E**rrors | Error rate | `rate(http_errors_total[5m]) / rate(http_requests_total[5m])` |
| **D**uration | Latency distribution | `histogram_quantile(0.99, rate(http_duration_bucket[5m]))` |

## 5. USE Method

Coined by Brendan Gregg — for **every resource** (CPU, memory, disk, network) monitor:

| Letter | Metric | Example |
|--------|--------|--------|
| **U**tilization | % time resource is busy | CPU busy % |
| **S**aturation | Extra work queued | Run queue length |
| **E**rrors | Error events | Disk errors/sec |

RED for request-driven services; USE for resource-constrained infrastructure.

```python
# Simulate RED dashboard for a hypothetical service
import random

random.seed(2024)

def simulate_service_metrics(n_samples: int = 60) -> dict:
    """Simulate 60 seconds of service metrics."""
    total_requests = 0
    total_errors   = 0
    latencies      = []

    for _ in range(n_samples):
        rps   = random.randint(80, 120)
        err   = random.randint(0, 3)
        total_requests += rps
        total_errors   += err
        for _ in range(rps):
            latencies.append(abs(random.gauss(50, 20)))  # ms

    latencies.sort()
    p99_idx = int(len(latencies) * 0.99)

    return {
        "rate_rps":    total_requests / n_samples,
        "error_rate":  total_errors   / total_requests,
        "p50_ms":      latencies[int(len(latencies) * 0.50)],
        "p99_ms":      latencies[p99_idx],
    }


metrics = simulate_service_metrics()
print("RED Dashboard")
print("=" * 30)
print(f"  Rate      : {metrics['rate_rps']:.1f} req/s")
print(f"  Error rate: {metrics['error_rate']*100:.2f}%")
print(f"  p50 latency: {metrics['p50_ms']:.1f} ms")
print(f"  p99 latency: {metrics['p99_ms']:.1f} ms")
```

## Key Takeaways

- Prometheus uses a **label-based dimensional model** — keep cardinality low
- **Counters** only go up; **Gauges** can go down; **Histograms** capture distributions
- **RED** (Rate/Errors/Duration) covers service health from a user perspective
- **USE** (Utilization/Saturation/Errors) covers resource health from an infrastructure perspective
- Always register metrics at startup — sparse data is better than missing data

## 🧠 Quick Quiz

**Q1.** A Prometheus Counter metric:
- A) Can increase and decrease
- B) **Only increases — tracks total events (requests, errors, etc.)** ✓
- C) Measures current values
- D) Stores distributions

**Q2.** A histogram metric is used for:
- A) Counting unique events
- B) Tracking the current value of a metric
- C) **Measuring distributions of values (e.g. request duration buckets)** ✓
- D) Comparing two time series

**Q3.** PromQL `rate(counter[5m])` computes:
- A) The total value over 5 minutes
- B) **The per-second increase rate averaged over 5 minutes** ✓
- C) The maximum value in 5 minutes
- D) The current counter value

<details><summary>Answers</summary>1-B, 2-C, 3-B</details>

## Hands-on Lab

A runnable companion lives in `./lab_02_metrics_and_monitoring/`. It packages a tiny FastAPI service that exposes the RED triad alongside Prometheus, Grafana, and Alertmanager.

```sh
cd lab_02_metrics_and_monitoring
make up           # build + start sample-app + prometheus + grafana + alertmanager
make load         # steady 10 rps against /work
make alert-test   # 30% error rate + 700ms mean latency -> HighErrorRate + HighLatencyP99 fire
make down         # stop the stack
```

Open Grafana at <http://localhost:3000> (admin/admin) and load **SE303 / RED Golden Signals — sample-app** to watch rate, error ratio, and p50/p95/p99 latency in real time. See the lab `README.md` for the full walk-through, PromQL queries, and teardown.
