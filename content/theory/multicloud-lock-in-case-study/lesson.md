Full walk-through of the architecture, sequencing, and instrumentation needed for a mid-stage B2B SaaS to deliberately leave single-cloud lock-in. Provider APIs are mocked as plain Python objects so every cell runs on stdlib alone; the same logic ports to Terraform / boto3 / google-cloud-* / azure-sdk by swapping the mocks for real clients.

**Scenario.** `acme-platform` runs entirely on AWS us-east-1: Django on EKS, Postgres on RDS, ML on SageMaker, identity scattered across IAM. Leadership signs off on a 12-month plan: keep the app and the database on AWS, move ML to GCP for cost and tooling, consolidate identity behind Azure AD. The job: a migration plan that holds availability, controls egress, and proves the case with numbers a CFO will read.

## Architecture under test

```
         users
           |
  [Azure AD: OIDC IdP]
           |    \
           v     \___ WIF ___> [GCP: Vertex AI, GCS Parquet]
   [AWS: EKS app + RDS Postgres]            ^
           |                                 |
           +------- nightly S3 -> GCS -------+
           |                                 |
           +------- cross-cloud inference ---+ (50 GB/mo egress)
```

Three boundaries, each with explicit costs:
1. **Identity boundary** — Azure AD as the single source of truth; AWS and GCP federate, no long-lived keys.
2. **Data boundary** — Parquet on object storage as the interchange; CSV-on-S3 never crosses a cloud.
3. **Compute boundary** — AWS for the transactional plane, GCP for the ML plane; the seam is the inference call.

Every boundary is a place failures compound. If a boundary's value does not exceed the operational tax it adds, delete it from the design.

## Part 1 — Mock the provider primitives

```python
import random, statistics, time
from dataclasses import dataclass, field
from typing import Callable, Optional

random.seed(13)

@dataclass(frozen=True)
class Workload:
    name: str
    provider: str            # 'aws' | 'gcp' | 'azure'
    compute_usd: float
    storage_usd: float
    egress_usd: float
    managed_usd: float

    @property
    def total(self) -> float:
        return self.compute_usd + self.storage_usd + self.egress_usd + self.managed_usd

CURRENT = [
    Workload('EC2 app servers (4 x c5.4xlarge)',          'aws', 2240,    0,    0,    0),
    Workload('RDS PostgreSQL Multi-AZ (r5.2xlarge)',      'aws',  980,  115,    0,    0),
    Workload('SageMaker training (ml.p3.2xlarge, 4h/d)',  'aws', 3600,    0,    0,  450),
    Workload('SageMaker inference endpoint',              'aws', 1200,    0,    0,    0),
    Workload('S3 training data + egress',                 'aws',    0, 11.5, 45.0,    0),
    Workload('Glue data prep',                            'aws',  320,    0,    0,    0),
]
TARGET = [
    Workload('EKS app servers (4 x m5.2xlarge)',          'aws', 1840,    0,    0,  120),
    Workload('RDS PostgreSQL Multi-AZ (r5.2xlarge)',      'aws',  980,  115,    0,    0),
    Workload('Vertex AI training (A100, 4h/day)',         'gcp', 1800,    0,    0,    0),
    Workload('Vertex AI online prediction endpoint',      'gcp',  600,    0,    0,    0),
    Workload('GCS feature store (Parquet, 500 GB)',       'gcp',    0, 10.0,    0,    0),
    Workload('S3 -> GCS nightly sync + egress',           'aws',    0, 11.5, 45.0,    0),
    Workload('App -> GCP inference egress (50 GB/mo)',    'aws',    0,    0,  4.5,    0),
    Workload('Azure AD (P1, 100 users)',                  'azure',  0,    0,    0,   60),
]
```

**Why typed workloads.** A `Workload` is a contract — every new line item is a code change a peer can review. Spreadsheets drift; typed lists do not.

## Part 2 — Cost diff with by-provider breakdown

```python
def cost_summary(items, label):
    by = {}
    for w in items:
        by[w.provider] = by.get(w.provider, 0) + w.total
    total = sum(w.total for w in items)
    print(f'=== {label} ===')
    for p in sorted(by):
        print(f'  {p:6s} ${by[p]:>9,.2f}/mo')
    print(f'  {"total":6s} ${total:>9,.2f}/mo\n')
    return total

cur = cost_summary(CURRENT, 'CURRENT (AWS-only)')
tgt = cost_summary(TARGET,  'TARGET (AWS + GCP + Azure ID)')
print(f'monthly delta: ${cur - tgt:,.2f}  ({(cur - tgt)/cur*100:.1f}% of current)')
print(f'annual delta:  ${(cur - tgt)*12:,.2f}')
print('note: operational complexity also rises; savings must justify it.')
```

**Read the breakdown, not the bottom line.** The headline savings are real, but the new `app -> GCP inference egress` line is the seam you will be paged about. Every multi-cloud design has at least one of these; they grow non-linearly with traffic and must be modelled before commitment, not after.

## Part 3 — Identity federation (Azure AD -> AWS + GCP, no long-lived keys)

```python
@dataclass(frozen=True)
class OIDCToken:
    subject: str
    issuer: str
    audience: str
    expires_in_s: int

@dataclass
class CloudCreds:
    provider: str
    access_token: str
    expires_in_s: int

def issue_oidc(subject: str, audience: str) -> OIDCToken:
    return OIDCToken(subject=subject, issuer='https://login.microsoftonline.com/acme', audience=audience, expires_in_s=900)

def aws_assume_role_with_web_identity(token: OIDCToken, role_arn: str) -> CloudCreds:
    assert token.audience == 'sts.amazonaws.com', 'audience mismatch'
    assert token.expires_in_s > 0,                'token expired'
    return CloudCreds(provider='aws', access_token=f'aws-sts:{role_arn}:{token.subject}', expires_in_s=3600)

def gcp_workload_identity_federation(token: OIDCToken, sa_email: str) -> CloudCreds:
    assert token.audience.startswith('//iam.googleapis.com/projects/'), 'audience mismatch'
    return CloudCreds(provider='gcp', access_token=f'gcp-wif:{sa_email}:{token.subject}', expires_in_s=3600)

tok_aws = issue_oidc('user:alice@acme.io', 'sts.amazonaws.com')
tok_gcp = issue_oidc('user:alice@acme.io', '//iam.googleapis.com/projects/acme-ml/locations/global/workloadIdentityPools/aad/providers/aad')
print(aws_assume_role_with_web_identity(tok_aws, 'arn:aws:iam::123:role/eks-app'))
print(gcp_workload_identity_federation(tok_gcp, 'vertex-runner@acme-ml.iam.gserviceaccount.com'))
```

**No long-lived keys, anywhere.** The federation flow above means there is exactly *one* trust root (Azure AD), exactly *one* place to disable a leaver, and zero stored secrets in either cloud. If you cannot draw the trust chain on a whiteboard, your federation design is wrong and you will discover it during an incident.

## Part 4 — Data portability: CSV-on-S3 -> Parquet on object storage

```python
@dataclass
class ParquetSync:
    src_uri: str           # s3://acme-train/2026/05/data.csv
    dst_uri: str           # gs://acme-train-parquet/2026/05/data.parquet
    bytes_transferred: int
    schema_checked: bool

def sync_csv_to_parquet(src_uri: str, dst_uri: str, gb: float, schema_ok: bool) -> ParquetSync:
    # Mock: validate the schema before write; refuse on drift.
    if not schema_ok:
        raise RuntimeError(f'schema drift detected at {src_uri}; refusing to sync')
    return ParquetSync(src_uri, dst_uri, int(gb * 1e9), True)

syncs = [
    sync_csv_to_parquet(f's3://acme-train/day-{d:02d}.csv', f'gs://acme-train-parquet/day-{d:02d}.parquet', 16.5, schema_ok=True)
    for d in range(1, 4)
]
for s in syncs:
    print(f'  {s.src_uri:40s} -> {s.dst_uri:50s}  {s.bytes_transferred/1e9:.1f} GB')

try:
    sync_csv_to_parquet('s3://acme-train/bad.csv', 'gs://acme-train-parquet/bad.parquet', 1.0, schema_ok=False)
except RuntimeError as e:
    print(f'\nrefused: {e}')
```

**Parquet is the interchange, not the storage.** Both clouds can read Parquet at full throughput; CSV is the format that costs you twice (size, schema drift). Sync jobs that do not validate schema *will* eventually propagate a corruption across the boundary.

## Part 5 — Strangler-fig cutover with a disagreement gate

```python
@dataclass
class InferenceReq:
    rid: str
    features: list[float]
    expected: int

def sagemaker_infer(req):
    pred = 1 if sum(req.features) > 2.5 else 0
    return pred, max(80, random.gauss(180, 30))

def vertex_infer(req):
    pred = 1 if sum(req.features) + random.gauss(0, 0.05) > 2.5 else 0
    return pred, max(20, random.gauss(45, 10))

def parallel_run(reqs, vertex_pct):
    sm_lat, v_lat = [], []
    sm_ok = v_ok = disagree = n_v = 0
    for r in reqs:
        sm_p, sm_t = sagemaker_infer(r); sm_lat.append(sm_t)
        if sm_p == r.expected: sm_ok += 1
        if random.random() < vertex_pct:
            n_v += 1
            v_p, v_t = vertex_infer(r); v_lat.append(v_t)
            if v_p == r.expected: v_ok += 1
            if v_p != sm_p:       disagree += 1
    return {
        'n': len(reqs), 'vertex_pct': vertex_pct, 'n_v': n_v,
        'sm_acc': sm_ok / len(reqs),
        'v_acc':  v_ok / n_v if n_v else None,
        'disagreement': disagree / n_v if n_v else None,
        'sm_p50_ms': sorted(sm_lat)[len(sm_lat)//2],
        'v_p50_ms':  sorted(v_lat)[len(v_lat)//2] if v_lat else None,
    }

def cutover_gate(m, max_disagree=0.005):
    if m['disagreement'] is None:                  return False, 'no vertex traffic yet'
    if m['disagreement'] > max_disagree:           return False, f'disagreement {m["disagreement"]:.3%} > {max_disagree:.3%}'
    return True, f'disagreement {m["disagreement"]:.3%}: ready'

REQS = [InferenceReq(f'r{i:04d}', [random.uniform(0,1) for _ in range(5)], random.choice([0,1])) for i in range(500)]
for label, pct in [('shadow 5%', 0.05), ('canary 20%', 0.20), ('ramp 50%', 0.50), ('cutover 100%', 1.00)]:
    random.seed(42)
    m = parallel_run(REQS, pct)
    ok, reason = cutover_gate(m)
    print(f'{label:14s} n_v={m["n_v"]:>3d}  sm_acc={m["sm_acc"]:.3f}  v_acc={m["v_acc"]:.3f}  '
          f'p50: sm={m["sm_p50_ms"]:.0f}ms v={m["v_p50_ms"]:.0f}ms  [{"GO" if ok else "HOLD"}] {reason}')
```

**Strangler-fig with a gate beats a flag-day, every time.** Shadow, canary, ramp, cutover — each phase has a numeric exit condition. The gate is not a vibe call; it is a Python function that returns a bool. Anything else and the cutover stalls on opinions in a meeting room.

## Part 6 — Resilience composition (active-passive across two clouds)

```python
@dataclass
class Endpoint:
    provider: str
    healthy: bool
    region: str

def dns_resolve(endpoints, prefer):
    primary = next((e for e in endpoints if e.provider == prefer and e.healthy), None)
    if primary: return primary
    return next((e for e in endpoints if e.healthy), None)

def availability_serial(a, b):
    return a * b  # both must be up: lower availability

def availability_parallel(a, b):
    return 1 - (1 - a) * (1 - b)  # either suffices: higher availability

endpoints = [Endpoint('aws', True, 'us-east-1'), Endpoint('gcp', True, 'us-central1')]
print(f'normal:                {dns_resolve(endpoints, "aws")}')
endpoints[0] = Endpoint('aws', False, 'us-east-1')
print(f'aws region down:       {dns_resolve(endpoints, "aws")}')

aws_slo = 0.9995; gcp_slo = 0.9990
print(f'\nserial composition (need both clouds):   {availability_serial(aws_slo, gcp_slo):.5f}')
print(f'parallel composition (either suffices):  {availability_parallel(aws_slo, gcp_slo):.5f}')
```

**Composition rule.** Adding a second cloud *in serial* (you need both up) makes you less available, not more. It only buys availability when the dependency is *parallel* — either cloud satisfies the request. Most multi-cloud designs start as parallel on paper and become serial in production (shared identity, shared data sync, shared CI). Re-audit yearly.

## Part 7 — Decision table: when each multi-cloud move is right

| Driver | When it is real | When it is a bad reason |
|---|---|---|
| Cost arbitrage | One workload class is materially cheaper on the other cloud (e.g. ML compute) | A spot-price gap that closes quarterly |
| Vendor leverage | Negotiating a new commit; you need a credible BATNA | You already signed the commit yesterday |
| Data residency | A jurisdiction your primary cloud does not serve | A regulation you assumed without reading |
| Resilience | Truly parallel failure domains (different orgs, different ops) | A serial composition you label "parallel" on the diagram |
| Team skill | A team that is already shipping on the other cloud | A hiring slide that says "polyglot" |
| Identity consolidation | Centralising on a corporate IdP you already pay for | Adding a third IdP just to be neutral |

Print this; tape it to the next architecture review.

## Part 8 — Regression gate and migration runbook

```python
BASELINE = {
    'monthly_usd':         8_700,  # post-migration target
    'egress_usd':            120,  # cross-cloud egress headroom
    'inference_p50_ms':       60,  # vertex p50 must beat sagemaker by 3x
    'identity_long_lived':     0,  # zero long-lived keys in either cloud
}
TOL = 0.10

current = {
    'monthly_usd':         sum(w.total for w in TARGET),
    'egress_usd':          sum(w.egress_usd for w in TARGET),
    'inference_p50_ms':    50,
    'identity_long_lived':  0,
}
failed = []
for k, base in BASELINE.items():
    cur = current[k]
    if k == 'identity_long_lived':
        ok = cur == 0
    elif k == 'inference_p50_ms' or k == 'egress_usd' or k == 'monthly_usd':
        ok = cur <= base * (1 + TOL)
    print(f'  {k:22s} baseline={base:>8}  current={cur:>8}  [{"OK" if ok else "FAIL"}]')
    if not ok: failed.append(k)
print(f'\nmigration verdict: {"BLOCK" if failed else "PROCEED"}')
```

**Migration runbook (one page).**

| Step | Owner | Exit condition |
|---|---|---|
| 1. Stand up Azure AD as IdP; configure AWS + GCP federation | Platform | OIDC + WIF working; zero long-lived keys |
| 2. Build the Parquet sync; validate schema on every run | Data | 7 nights without a refusal |
| 3. Stand up Vertex AI endpoint behind a feature flag | ML | Shadow traffic at 5% for 7 days |
| 4. Run the parallel-run gate at 5 / 20 / 50 / 100% | ML + Platform | Disagreement < 0.5% at each phase |
| 5. Decommission SageMaker | ML | Endpoint deleted, IAM role revoked |
| 6. Re-run the cost model on 30 days of live data | Finance + Platform | Within 10% of target |
| 7. Quarterly: re-audit the serial-vs-parallel resilience map | Architecture | Map matches reality |

Anything not in the runbook is not part of the migration. New asks go to the next quarter, or they re-open the runbook with an owner and an exit condition.

## Reflection

**What this case study should leave you with**

- Multi-cloud is *deliberate boundaries*, not a default posture. Each boundary has an operational tax that must be paid by a specific value.
- Egress is the seam that bites; model it before commitment, instrument it after.
- Federation gives you one trust root and zero long-lived keys; everything else is a footgun waiting for an incident.
- Parquet (or any open columnar format) is what makes the next migration cheap; CSV makes it expensive twice.
- A strangler-fig cutover with a numeric disagreement gate beats a flag-day every time.
- Adding a second cloud in *serial* lowers availability; only *parallel* composition raises it. Audit yearly.



**Where to go next.** `CLD303 Cloud Cost Engineering` deepens the egress and commitment modelling; `SYS401 Reliability Engineering` formalises the serial-vs-parallel availability calculus; `AI704 Open Source LLMs` is the *next* candidate workload to lift off the proprietary inference vendor — same playbook, different boundary.
