Topics: trace/span model · W3C Trace Context · in-process tracer · sampling strategies

## 1. The Trace / Span Model

A **trace** represents the end-to-end journey of a request through a distributed system.

A **span** is a single unit of work within a trace:
- Has a `trace_id` (shared across all spans in one request)
- Has a `span_id` (unique to this operation)
- Has a `parent_span_id` (nil for the root span)
- Records `start_time`, `end_time`, `service`, `operation`, `status`, and optional `attributes`

```
Trace abc123
├── [span-1] api-gateway  POST /checkout     0ms → 210ms
│   ├── [span-2] order-svc  create_order    10ms → 100ms
│   │   └── [span-3] postgres  INSERT orders 15ms → 40ms
│   └── [span-4] payment-svc charge_card    110ms → 200ms
```

```python
import time
import uuid
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any
from contextlib import contextmanager


@dataclass
class Span:
    trace_id:      str
    span_id:       str
    parent_span_id: Optional[str]
    service:       str
    operation:     str
    start_ns:      int
    end_ns:        int = 0
    status:        str = "ok"       # ok | error
    attributes:    Dict[str, Any] = field(default_factory=dict)

    @property
    def duration_ms(self) -> float:
        return (self.end_ns - self.start_ns) / 1_000_000

    def finish(self) -> None:
        self.end_ns = time.monotonic_ns()

    def set_error(self, message: str) -> None:
        self.status = "error"
        self.attributes["error.message"] = message


# Create a standalone span
s = Span(
    trace_id="abc123",
    span_id=str(uuid.uuid4())[:8],
    parent_span_id=None,
    service="api-gateway",
    operation="POST /checkout",
    start_ns=time.monotonic_ns(),
)
time.sleep(0.01)  # simulate 10ms work
s.finish()
print(f"Span: {s.operation}  [{s.duration_ms:.2f} ms]  status={s.status}")
```

## 2. W3C Trace Context

The [W3C Trace Context](https://www.w3.org/TR/trace-context/) standard defines HTTP headers for context propagation:

```
traceparent: 00-{trace_id_hex_32}-{parent_id_hex_16}-{flags}
tracestate:  vendor1=value1,vendor2=value2
```

Example:
```
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
```
- Version: `00`
- Trace ID: `4bf92f3577b34da6a3ce929d0e0e4736` (128-bit)
- Parent Span ID: `00f067aa0ba902b7` (64-bit)
- Flags: `01` = sampled

```python
import secrets


def new_trace_id() -> str:
    return secrets.token_hex(16)  # 32 hex chars = 128 bits


def new_span_id() -> str:
    return secrets.token_hex(8)   # 16 hex chars = 64 bits


def make_traceparent(trace_id: str, span_id: str, sampled: bool = True) -> str:
    flags = "01" if sampled else "00"
    return f"00-{trace_id}-{span_id}-{flags}"


def parse_traceparent(header: str) -> Dict[str, str]:
    parts = header.split("-")
    if len(parts) != 4 or parts[0] != "00":
        raise ValueError(f"Invalid traceparent: {header}")
    return {
        "version":  parts[0],
        "trace_id": parts[1],
        "span_id":  parts[2],
        "flags":    parts[3],
        "sampled":  parts[3] == "01",
    }


tid = new_trace_id()
sid = new_span_id()
header = make_traceparent(tid, sid)
print(f"traceparent header: {header}")

parsed = parse_traceparent(header)
for k, v in parsed.items():
    print(f"  {k}: {v}")
```

## 3. In-Process Tracer Implementation

```python
class Tracer:
    """Minimal in-process tracer — collects spans in memory."""

    def __init__(self, service: str):
        self.service = service
        self._spans: List[Span] = []
        self._active_span: Optional[Span] = None

    @contextmanager
    def start_span(self, operation: str, attributes: Optional[Dict] = None):
        parent_id = self._active_span.span_id if self._active_span else None
        trace_id  = self._active_span.trace_id if self._active_span else new_trace_id()

        span = Span(
            trace_id=trace_id,
            span_id=new_span_id(),
            parent_span_id=parent_id,
            service=self.service,
            operation=operation,
            start_ns=time.monotonic_ns(),
            attributes=attributes or {},
        )

        previous = self._active_span
        self._active_span = span
        try:
            yield span
        except Exception as exc:
            span.set_error(str(exc))
            raise
        finally:
            span.finish()
            self._spans.append(span)
            self._active_span = previous

    def dump(self) -> None:
        print(f"\nTrace dump — service={self.service} ({len(self._spans)} spans)")
        for sp in self._spans:
            indent = "  " if sp.parent_span_id else ""
            print(f"  {indent}{sp.operation:<35} {sp.duration_ms:6.2f} ms  [{sp.status}]")


tracer = Tracer("checkout-service")

with tracer.start_span("POST /checkout") as root:
    with tracer.start_span("validate_cart"):
        time.sleep(0.005)
    with tracer.start_span("db.query orders"):
        time.sleep(0.012)
    with tracer.start_span("payment.charge"):
        time.sleep(0.020)

tracer.dump()
```

## 4. Sampling Strategies

Tracing every request at high volume is expensive. Sampling reduces cost while preserving signal.

| Strategy | How | Pros | Cons |
|----------|-----|------|------|
| **Head-based** | Decide at trace start (e.g. 1%) | Simple, low overhead | May miss rare errors |
| **Tail-based** | Decide after trace completes | Can keep all errors | Requires buffering full traces |
| **Adaptive** | Adjust rate based on traffic | Best coverage | Complex |
| **Error-always** | Always sample if error occurred | Never miss incidents | Slightly complex |

```python
import random as _rand


class HeadBasedSampler:
    """Sample a fixed percentage of traces at ingestion time."""

    def __init__(self, rate: float = 0.01):
        self.rate = rate

    def should_sample(self) -> bool:
        return _rand.random() < self.rate


class TailBasedSampler:
    """Buffer completed traces; decide based on outcome."""

    def __init__(self, base_rate: float = 0.01):
        self.base_rate = base_rate

    def should_keep(self, spans: List[Span]) -> bool:
        # Always keep traces with errors
        if any(s.status == "error" for s in spans):
            return True
        # Always keep slow traces (>500ms root)
        root = next((s for s in spans if not s.parent_span_id), None)
        if root and root.duration_ms > 500:
            return True
        # Otherwise probabilistic
        return _rand.random() < self.base_rate


# Simulate sampling decisions
_rand.seed(42)
head = HeadBasedSampler(rate=0.05)
head_kept = sum(1 for _ in range(10_000) if head.should_sample())
print(f"Head-based sampler (5%): kept {head_kept} / 10000 traces ({head_kept/100:.1f}%)")
```

## Key Takeaways

- A **trace** is a DAG of **spans** sharing one `trace_id`
- Use **W3C traceparent** headers to propagate context across HTTP calls
- Implement tracing as a **context manager** to guarantee span completion
- **Tail-based sampling** is more accurate but requires buffering — prefer it when errors must never be missed
- Always record `error` status on spans — it makes root cause analysis instant

## 🧠 Quick Quiz

**Q1.** A distributed trace captures:
- A) CPU usage across all services
- B) **The full journey of a request through multiple services as spans** ✓
- C) Network packet routing
- D) Database query plans

**Q2.** A span in distributed tracing represents:
- A) The entire request lifetime
- B) A network connection
- C) **A single unit of work (e.g. one service call, one DB query)** ✓
- D) An error event

**Q3.** OpenTelemetry provides:
- A) A proprietary tracing format
- B) **A vendor-neutral standard for traces, metrics, and logs** ✓
- C) Only distributed tracing
- D) A hosted observability platform

<details><summary>Answers</summary>1-B, 2-C, 3-B</details>
