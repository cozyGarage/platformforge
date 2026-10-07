## Implementing Service Level Objectives, Error Budgets, and Multi-Window Alerting

The Google SRE model centres on Service Level Objectives (SLOs): quantified reliability targets that define the acceptable failure rate for a service. When the error budget (the allowed failures within the SLO) burns faster than expected, alerts fire and reliability work takes priority over feature work. This case study implements a complete SLO alerting system: request-level SLI measurement, error budget calculation, and the multi-window burn rate alerting algorithm from the Google SRE Workbook.

## Part 1: SLI Measurement and Error Budget Calculation

We generate synthetic request-level telemetry data, implement a rolling-window SLI aggregator, and calculate the error budget consumption rate. The error budget is not just a metric — it is the shared language between product and engineering that determines when reliability work takes precedence.

```python
import math
import random
import statistics
from dataclasses import dataclass
from datetime import datetime, timedelta

random.seed(42)

# ============================================================
# SLO Configuration
# ============================================================

@dataclass(frozen=True)
class SLO:
    name: str
    target: float        # e.g. 0.999 = 99.9%
    window_days: int     # Rolling window (typically 30 days)
    description: str

AVAILABILITY_SLO = SLO(
    name="api_availability",
    target=0.999,   # 99.9% — allows 43.2 minutes/month downtime
    window_days=30,
    description="Fraction of requests returning non-5xx status codes",
)

LATENCY_SLO = SLO(
    name="api_latency_p99",
    target=0.95,    # 95% of requests under 300ms
    window_days=30,
    description="Fraction of requests with response time < 300ms",
)

def error_budget_minutes(slo: SLO) -> float:
    """Total error budget in minutes for the SLO window."""
    return (1 - slo.target) * slo.window_days * 24 * 60

print(f"SLO Configuration:")
print(f"  {AVAILABILITY_SLO.name}: {AVAILABILITY_SLO.target*100:.1f}% target")
print(f"    Error budget: {error_budget_minutes(AVAILABILITY_SLO):.1f} minutes / {AVAILABILITY_SLO.window_days} days")
print(f"  {LATENCY_SLO.name}: {LATENCY_SLO.target*100:.1f}% target")
print(f"    Error budget: {error_budget_minutes(LATENCY_SLO):.1f} minutes / {LATENCY_SLO.window_days} days")

# ============================================================
# Synthetic request telemetry
# ============================================================

@dataclass
class Request:
    timestamp: datetime
    status_code: int
    response_time_ms: float
    endpoint: str

def generate_telemetry(hours: int = 72, rps: int = 100) -> list[Request]:
    """
    Generate synthetic request log.
    Includes a 4-hour incident window with elevated error rate.
    """
    start = datetime(2024, 3, 1, 0, 0, 0)
    requests = []
    
    INCIDENT_START = start + timedelta(hours=48)  # Incident at hour 48
    INCIDENT_END   = INCIDENT_START + timedelta(hours=4)
    
    # Sample once per minute (scaled down for tractability)
    for minute in range(hours * 60):
        ts = start + timedelta(minutes=minute)
        n_requests = max(1, int(rps/60 + random.gauss(0, 2)))
        
        in_incident = INCIDENT_START <= ts < INCIDENT_END
        
        for _ in range(n_requests):
            error_prob = 0.05 if in_incident else 0.0005  # 5% vs 0.05% error rate
            is_error = random.random() < error_prob
            
            latency = (
                random.lognormvariate(5.0, 0.5) if in_incident  # Slower during incident
                else random.lognormvariate(4.5, 0.4)
            )
            
            requests.append(Request(
                timestamp=ts,
                status_code=500 if is_error else random.choice([200, 200, 200, 200, 201, 304]),
                response_time_ms=latency,
                endpoint=random.choice(["/api/users", "/api/orders", "/api/products", "/health"]),
            ))
    
    return requests

telemetry = generate_telemetry(hours=72)
n_errors = sum(1 for r in telemetry if r.status_code >= 500)
n_slow   = sum(1 for r in telemetry if r.response_time_ms > 300)
print(f"\nGenerated {len(telemetry):,} requests over 72 hours")
print(f"  Error rate  : {n_errors/len(telemetry)*100:.3f}% ({n_errors:,} errors)")
print(f"  Slow rate   : {n_slow/len(telemetry)*100:.2f}% ({n_slow:,} > 300ms)")

# ============================================================
# SLI measurement over rolling windows
# ============================================================

def compute_sli_window(requests: list[Request], 
                        start: datetime, end: datetime,
                        slo: SLO) -> dict:
    """Compute SLI for a time window."""
    window_requests = [r for r in requests if start <= r.timestamp < end]
    
    if not window_requests:
        return {"sli": 1.0, "total": 0, "bad": 0, "window_hours": (end-start).seconds//3600}
    
    if slo.name == "api_availability":
        bad = sum(1 for r in window_requests if r.status_code >= 500)
    elif slo.name == "api_latency_p99":
        bad = sum(1 for r in window_requests if r.response_time_ms > 300)
    else:
        bad = 0
    
    total = len(window_requests)
    sli = 1.0 - (bad / total)
    
    return {
        "sli": sli,
        "total": total,
        "bad": bad,
        "error_rate": bad / total,
        "window_hours": (end - start).total_seconds() / 3600,
    }

# Compute hourly SLI over the 72-hour window
base_time = datetime(2024, 3, 1, 0, 0, 0)
hourly_slis = []
for hour in range(72):
    window_start = base_time + timedelta(hours=hour)
    window_end   = window_start + timedelta(hours=1)
    sli_data = compute_sli_window(telemetry, window_start, window_end, AVAILABILITY_SLO)
    hourly_slis.append({"hour": hour, **sli_data})

# Error budget consumption over time
TARGET = AVAILABILITY_SLO.target
total_requests_72h = sum(h["total"] for h in hourly_slis)
total_bad_72h      = sum(h["bad"] for h in hourly_slis)
overall_sli = 1 - total_bad_72h / total_requests_72h

# Scale error budget for 72-hour window (from 30-day SLO)
budget_requests_72h = total_requests_72h * (1 - TARGET)
budget_consumed = total_bad_72h / budget_requests_72h if budget_requests_72h > 0 else 0.0

print(f"\nSLI measurement (72-hour window):")
print(f"  Total requests: {total_requests_72h:,}")
print(f"  Total errors  : {total_bad_72h:,}")
print(f"  Overall SLI   : {overall_sli:.5f}")
print(f"  SLO target    : {TARGET:.3f}")
print(f"  SLO met       : {overall_sli >= TARGET}")
print(f"  Error budget consumed: {budget_consumed:.1%} (of 72-hour allotment)")

# Show hourly SLI around the incident
print(f"\nHourly SLI around incident (hours 46–52):")
print(f"{'Hour':>6s} {'SLI':>8s} {'Errors':>8s} {'Total':>8s} {'Budget hit':>12s}")
print("-" * 47)
for h in hourly_slis[46:53]:
    budget_hit = h['sli'] < TARGET
    print(f"  {h['hour']:>4d} {h['sli']:>8.5f} {h['bad']:>8d} {h['total']:>8d} {'YES' if budget_hit else 'no':>12s}")
```

### Key Takeaway — Part 1

The error budget translates the abstract SLO percentage into a concrete quantity that can be consumed and tracked. During normal operations, 72 hours should consume approximately (72 / (30*24)) = 10% of the monthly error budget. The incident at hours 48–52 concentrates error budget consumption dramatically — this asymmetric consumption pattern is exactly what the multi-window burn rate alert is designed to detect quickly.

## Part 2: Multi-Window Burn Rate Alerting

The burn rate measures how quickly the error budget is being consumed relative to the expected rate. A burn rate of 1.0 means the budget is consumed at exactly the rate that would exhaust it at the end of the SLO window; a burn rate of 14.4x means the budget would be exhausted in ~2 hours. The Google SRE Workbook recommends a two-window alerting strategy: a short window catches fast burns, a long window prevents false positives from brief spikes.

```python
# ============================================================
# Multi-window burn rate alerting
# ============================================================

def compute_burn_rate(requests: list[Request], 
                       window_start: datetime, 
                       window_hours: float,
                       slo: SLO) -> dict:
    """
    Compute error budget burn rate for a time window.
    Burn rate = (actual error rate) / (allowed error rate)
    A burn rate > 1.0 means the budget is being exhausted faster than planned.
    """
    window_end = window_start + timedelta(hours=window_hours)
    sli_data = compute_sli_window(requests, window_start, window_end, slo)
    
    allowed_error_rate = 1 - slo.target
    actual_error_rate = sli_data["error_rate"]
    
    burn_rate = actual_error_rate / allowed_error_rate if allowed_error_rate > 0 else 0.0
    
    # Time-to-exhaustion at current burn rate
    if burn_rate > 0:
        time_to_exhaustion_days = slo.window_days / burn_rate
    else:
        time_to_exhaustion_days = float('inf')
    
    return {
        "window_hours": window_hours,
        "burn_rate": burn_rate,
        "sli": sli_data["sli"],
        "error_rate": actual_error_rate,
        "time_to_exhaustion_days": time_to_exhaustion_days,
        "total_requests": sli_data["total"],
        "bad_requests": sli_data["bad"],
    }


# Alert thresholds from Google SRE Workbook (Table 5-2)
# Two-window strategy: short window catches fast burns, long window avoids false positives
ALERT_CONFIG = [
    {"name": "P1_CRITICAL",  "short_hours": 1,   "long_hours": 5,   "burn_rate_threshold": 14.4},
    {"name": "P2_HIGH",      "short_hours": 6,   "long_hours": 30,  "burn_rate_threshold": 6.0},
    {"name": "P3_MEDIUM",    "short_hours": 24,  "long_hours": 72,  "burn_rate_threshold": 3.0},
    {"name": "P4_SLOW_BURN", "short_hours": 72,  "long_hours": 360, "burn_rate_threshold": 1.0},
]

def check_alerts(requests: list[Request], check_time: datetime, slo: SLO) -> list[dict]:
    """Evaluate all alert conditions at a given point in time."""
    fired_alerts = []
    
    for alert in ALERT_CONFIG:
        short_start = check_time - timedelta(hours=alert["short_hours"])
        long_start  = check_time - timedelta(hours=alert["long_hours"])
        
        short_br = compute_burn_rate(requests, short_start, alert["short_hours"], slo)
        long_br  = compute_burn_rate(requests, long_start,  alert["long_hours"],  slo)
        
        # Alert fires only when BOTH windows show high burn rate
        fires = (short_br["burn_rate"] >= alert["burn_rate_threshold"] and
                 long_br["burn_rate"]  >= alert["burn_rate_threshold"])
        
        fired_alerts.append({
            "alert": alert["name"],
            "fires": fires,
            "short_burn": round(short_br["burn_rate"], 2),
            "long_burn":  round(long_br["burn_rate"],  2),
            "threshold":  alert["burn_rate_threshold"],
            "time_to_exhaustion": round(short_br["time_to_exhaustion_days"] * 24, 1),
        })
    
    return fired_alerts

# Evaluate alerts at different points: before, during, and after incident
base_time = datetime(2024, 3, 1, 0, 0, 0)
INCIDENT_START = base_time + timedelta(hours=48)

check_points = [
    ("Pre-incident (hour 24)",    base_time + timedelta(hours=24)),
    ("Incident +1h (hour 49)",    base_time + timedelta(hours=49)),
    ("Incident +2h (hour 50)",    base_time + timedelta(hours=50)),
    ("Incident +4h (hour 52)",    base_time + timedelta(hours=52)),
    ("Post-incident (hour 56)",   base_time + timedelta(hours=56)),
]

for label, check_time in check_points:
    alerts = check_alerts(telemetry, check_time, AVAILABILITY_SLO)
    fired = [a for a in alerts if a["fires"]]
    print(f"\n{label}:")
    print(f"  {'Alert':15s} {'Short burn':>12s} {'Long burn':>10s} {'Threshold':>10s} {'TTX (hrs)':>10s} {'Status':>8s}")
    print(f"  {'-'*65}")
    for a in alerts:
        status = "🚨 FIRES" if a["fires"] else "ok"
        ttx = f"{a['time_to_exhaustion']:.1f}h" if a['time_to_exhaustion'] < 1000 else "∞"
        print(f"  {a['alert']:15s} {a['short_burn']:>12.2f}x {a['long_burn']:>10.2f}x "
              f"{a['threshold']:>10.1f}x {ttx:>10s} {status:>8s}")

# Error budget status report
print(f"\n\nERROR BUDGET STATUS REPORT")
print("=" * 50)
window_30d_start = base_time
window_30d_end   = base_time + timedelta(hours=72)  # Only 72h of data
total_req = len(telemetry)
total_err = sum(1 for r in telemetry if r.status_code >= 500)
sli_72h = 1 - total_err / total_req
budget_fraction_consumed = total_err / (total_req * (1 - AVAILABILITY_SLO.target)) if total_req > 0 else 0

print(f"  SLO target            : {AVAILABILITY_SLO.target*100:.1f}%")
print(f"  Measured SLI (72h)    : {sli_72h*100:.3f}%")
print(f"  SLO status            : {'MET' if sli_72h >= AVAILABILITY_SLO.target else 'BREACHED'}")
print(f"  Error budget consumed : {budget_fraction_consumed:.1%} (72h window)")
print(f"  Remaining budget      : {max(0, 1-budget_fraction_consumed):.1%}")
print()
print("Recommendation: " + (
    "Budget healthy — feature work can proceed" if budget_fraction_consumed < 0.5
    else "Budget depleted — reliability work should take priority"
))
```

### Key Takeaway — Part 2

The two-window alerting strategy eliminates two failure modes: a single-window alert would fire on brief spikes (false positives) or miss slow burns that don't exceed the threshold in any single window. The `short_burn AND long_burn` condition means: the service is currently burning fast (short window) AND has been burning fast for long enough that it's not just a transient spike (long window). The time-to-exhaustion metric converts an abstract burn rate into an actionable timeline: when TTX is 2 hours, the on-call engineer needs to act immediately; when TTX is 60 hours, a next-business-day response may be sufficient.

## Reflection Questions

1. The alerting thresholds (14.4x, 6x, 3x) in the Google SRE Workbook are derived from specific error budget consumption targets (5% in 1 hour, 10% in 6 hours, etc.). Derive the 14.4x threshold mathematically: if an incident burns 5% of a 30-day budget in 1 hour, what burn rate does that imply?
2. The two-window alerting strategy has a specific recall/precision trade-off. A single short window has high recall (catches all fast burns) but low precision (many false positives). How does adding the long-window condition change this trade-off?
3. This implementation computes burn rates on each check using the full request log. For a production system receiving millions of requests, how would you redesign the SLI computation to be efficient — specifically, what pre-aggregated time-series structure would enable O(1) window queries?
4. The SLO in this case study is binary: 99.9% of requests are good. Some SLOs are multi-dimensional (e.g., 99.9% availability AND 95% of requests under 300ms). How would you define the error budget when multiple SLO dimensions must be satisfied simultaneously?

## Going Deeper

- **Google SRE Workbook Ch. 5**: "Alerting on SLOs" — the chapter this case study implements, including the mathematical derivation of all threshold values and the evaluation of single-window vs multi-window strategies.
- **OpenTelemetry metrics documentation**: Study `Counter` and `Histogram` instrument types — the OpenTelemetry equivalents of the request-level telemetry collected here, with automatic aggregation by time window.
- **Prometheus recording rules**: Implement the burn rate calculation as a Prometheus recording rule — the production approach that pre-computes burn rates as time-series for efficient alerting evaluation.
- **Practical exercise**: Implement a "budget burn analysis" function that takes the request log and returns a breakdown of error budget consumption by endpoint, HTTP method, and hour of day. Identify which endpoint/time combination is responsible for the majority of budget consumption during the incident.
