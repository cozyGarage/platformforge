Topics: structured vs unstructured logs · JSON logging · correlation IDs · log levels · aggregation patterns

## 1. Unstructured vs Structured Logs

**Unstructured** (plain text):
```
2024-01-15 10:23:41 ERROR Failed to connect to DB: timeout after 5s
```
Hard to query at scale — requires fragile regex parsing.

**Structured** (JSON):
```json
{"ts":"2024-01-15T10:23:41Z","level":"error","msg":"db_connection_failed",
 "error":"timeout","duration_ms":5000,"service":"checkout"}
```
Machine-readable, filterable by any field, aggregatable across services.

Modern log platforms (Loki, Elasticsearch, Datadog) index structured fields natively.

```python
import json
import sys
import uuid
import datetime
from typing import Any, Dict, List, Optional


class StructuredLogger:
    """Minimal structured logger emitting JSON to stdout."""

    LEVELS = {"debug": 10, "info": 20, "warning": 30, "error": 40, "critical": 50}

    def __init__(self, service: str, min_level: str = "info"):
        self.service   = service
        self.min_level = self.LEVELS[min_level]
        self._context: Dict[str, Any] = {}

    def with_context(self, **kwargs: Any) -> "StructuredLogger":
        """Return a child logger with additional bound fields."""
        child = StructuredLogger(self.service)
        child.min_level = self.min_level
        child._context  = {**self._context, **kwargs}
        return child

    def _emit(self, level: str, msg: str, **kwargs: Any) -> None:
        if self.LEVELS[level] < self.min_level:
            return
        record = {
            "ts":      datetime.datetime.utcnow().isoformat() + "Z",
            "level":   level,
            "service": self.service,
            "msg":     msg,
            **self._context,
            **kwargs,
        }
        print(json.dumps(record))

    def debug(self,    msg: str, **kw) -> None: self._emit("debug",    msg, **kw)
    def info(self,     msg: str, **kw) -> None: self._emit("info",     msg, **kw)
    def warning(self,  msg: str, **kw) -> None: self._emit("warning",  msg, **kw)
    def error(self,    msg: str, **kw) -> None: self._emit("error",    msg, **kw)
    def critical(self, msg: str, **kw) -> None: self._emit("critical", msg, **kw)


log = StructuredLogger("order-service")
log.info("server_started", port=8080)
log.warning("high_memory_usage", rss_mb=780, threshold_mb=800)
log.error("payment_failed", order_id="ord-123", error="gateway_timeout", duration_ms=3001)
```

## 2. Log Levels

| Level | Numeric | When to use |
|-------|---------|-------------|
| DEBUG | 10 | Detailed diagnostics — disabled in production |
| INFO | 20 | Normal lifecycle events (startup, requests) |
| WARNING | 30 | Unexpected but handled (retry, fallback) |
| ERROR | 40 | Failure requiring attention (but service still running) |
| CRITICAL | 50 | System-level failure, imminent crash |

**Anti-patterns**:
- Log at ERROR for every 4xx (those are client errors, not service errors)
- Log DEBUG in production — massive volume, performance impact
- Inconsistent level semantics across services

```python
# Demonstrate level filtering
prod_log = StructuredLogger("api-gateway", min_level="warning")

prod_log.debug("request_parsed", path="/health")   # suppressed
prod_log.info("request_handled", status=200)        # suppressed
prod_log.warning("rate_limit_approaching", rps=950, limit=1000)
prod_log.error("upstream_unavailable", upstream="inventory-service")
```

## 3. Correlation IDs

In distributed systems, a single user action may span many services.  
A **correlation ID** (also called trace ID or request ID) ties all log lines from one request together.

Convention:
1. Edge service generates a UUID on entry (`X-Request-ID` header)
2. Every downstream service propagates it in both HTTP headers and log fields
3. Query `request_id = "abc-123"` in your log platform to see the full chain

```python
import uuid


def handle_request(request_id: Optional[str] = None) -> None:
    """Simulate multi-service request flow with correlation ID."""
    rid = request_id or str(uuid.uuid4())

    # Each service uses a child logger bound to this request
    gateway_log    = StructuredLogger("api-gateway").with_context(request_id=rid)
    order_log      = StructuredLogger("order-svc").with_context(request_id=rid)
    inventory_log  = StructuredLogger("inventory-svc").with_context(request_id=rid)

    gateway_log.info("request_received",   method="POST", path="/orders")
    order_log.info("order_created",        order_id="ord-999", user_id="u-42")
    inventory_log.info("stock_reserved",   sku="SKU-007", qty=1)
    order_log.info("payment_initiated",    gateway="stripe")
    gateway_log.info("response_sent",      status=201, duration_ms=142)


print("=== Correlated request log ===")
handle_request(request_id="req-demo-abc123")
```

## 4. Log Aggregation Patterns

Common architectures for collecting logs at scale:

```
Service → Sidecar agent (Fluentbit/Vector)
       → Message bus (Kafka / Cloud Pub/Sub)
       → Storage + Index (Elasticsearch / Loki)
       → Query UI (Grafana / Kibana)
```

Key concerns:
- **Buffering** — handle log bursts without dropping events
- **Backpressure** — slow consumers should not crash producers
- **Sampling** — at very high volume, sample DEBUG/INFO (never ERROR)
- **Retention** — short for debug logs, long for audit/compliance

```python
# Simple in-memory log buffer with sampling
from collections import deque
import random as _random


class SamplingLogBuffer:
    """Buffer that samples low-level logs, always keeps errors."""

    def __init__(self, max_size: int = 1000, sample_rate: float = 0.1):
        self.max_size    = max_size
        self.sample_rate = sample_rate
        self._buffer: deque = deque(maxlen=max_size)
        self.stats = {"accepted": 0, "sampled_out": 0}

    def write(self, record: Dict[str, Any]) -> None:
        level = record.get("level", "info")
        # Always accept warnings and above
        if level in ("warning", "error", "critical"):
            self._buffer.append(record)
            self.stats["accepted"] += 1
        elif _random.random() < self.sample_rate:
            record = {**record, "sampled": True}
            self._buffer.append(record)
            self.stats["accepted"] += 1
        else:
            self.stats["sampled_out"] += 1

    def flush(self) -> List[Dict]:
        records = list(self._buffer)
        self._buffer.clear()
        return records


buf = SamplingLogBuffer(sample_rate=0.1)
_random.seed(1)

for i in range(1000):
    level = _random.choice(["debug", "info", "info", "warning", "error"])
    buf.write({"level": level, "msg": f"event_{i}"})

print(f"Accepted  : {buf.stats['accepted']}")
print(f"Sampled out: {buf.stats['sampled_out']}")
print(f"Buffer size: {len(buf._buffer)}")
```

## Key Takeaways

- Structured (JSON) logs are **machine-readable** — prefer them always
- Use **correlation IDs** to link log lines across services for a single request
- Set **log levels** deliberately — ERROR logs should wake someone up
- At scale, **sample** high-volume low-severity logs; never sample errors
- Include actionable fields: `request_id`, `user_id`, `duration_ms`, `error`

## 🧠 Quick Quiz

**Q1.** Structured logs are preferable to plain text because:
- A) They are shorter
- B) **They are machine-parseable, enabling filtering, aggregation, and alerting** ✓
- C) They contain more information
- D) They are faster to write

**Q2.** A correlation ID in logs:
- A) Identifies the log level
- B) **Links all log entries across services for a single request** ✓
- C) Measures request latency
- D) Identifies the logging framework

**Q3.** Log sampling (not logging every event) is used to:
- A) Improve log quality
- B) Ensure all events are captured
- C) **Reduce storage costs and noise at high traffic volumes** ✓
- D) Speed up the application

<details><summary>Answers</summary>1-B, 2-B, 3-C</details>
