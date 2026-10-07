**Course:** DS301 — Databases & SQL

**Scenario.** It is Monday morning. The finance team’s weekly revenue dashboard query — the one feeding `revenue_by_region_week` — has gone from 800 ms to 47 seconds since Friday’s release. The schema didn’t change. The data volume grew by 4%. The on-call rotation is you.

This case study walks the full tuning loop: **profile → read the plan → pick the right index → weigh denormalisation → add a regression gate so the fix doesn’t silently rot.** Every step runs on stdlib `sqlite3` so you can adapt the workflow to any RDBMS (Postgres `EXPLAIN ANALYZE`, MySQL `EXPLAIN FORMAT=JSON`, SQL Server `SET STATISTICS IO`).

## Architecture under test

```
  dashboard ---> [revenue_by_region_week query] ---> result set
                          |
                          v
                  fact_orders (large)  JOIN  dim_date, dim_customer, dim_product (small)
                          |
                          +--- predicate: order_date BETWEEN ? AND ?
                          +--- group:     region, iso_week
                          +--- aggregate: SUM(net_revenue)
```

Tuning is a four-question loop. Skip a step and you optimise the wrong thing:
1. **Where is time going?** Wall-clock alone lies; you need row counts per plan node.
2. **What plan is the engine choosing?** `EXPLAIN`/`EXPLAIN ANALYZE` is the source of truth.
3. **What’s the cheapest structural fix?** Index, rewrite, materialised view, denormalisation — in that order of cost.
4. **How do we keep it fast?** A regression gate on plan shape and latency, not just on result correctness.

## Part 1 — Reproduce the slow query

```python
import sqlite3, random, time
from dataclasses import dataclass
from datetime import date, timedelta

random.seed(42)
conn = sqlite3.connect(':memory:')
conn.row_factory = sqlite3.Row
cur = conn.cursor()

cur.executescript("""
CREATE TABLE dim_date (
    date_key   INTEGER PRIMARY KEY,
    full_date  TEXT,
    iso_week   INTEGER,
    month      INTEGER,
    quarter    INTEGER
);
CREATE TABLE dim_customer (
    customer_key INTEGER PRIMARY KEY,
    region       TEXT,
    segment      TEXT
);
CREATE TABLE dim_product (
    product_key INTEGER PRIMARY KEY,
    category    TEXT,
    unit_price  REAL
);
CREATE TABLE fact_orders (
    order_key    INTEGER PRIMARY KEY,
    date_key     INTEGER REFERENCES dim_date(date_key),
    customer_key INTEGER REFERENCES dim_customer(customer_key),
    product_key  INTEGER REFERENCES dim_product(product_key),
    quantity     INTEGER,
    net_revenue  REAL
);
""")

base = date(2024, 1, 1)
for i in range(365):
    d = base + timedelta(days=i)
    cur.execute('INSERT INTO dim_date VALUES (?,?,?,?,?)',
                (d.toordinal(), d.isoformat(), d.isocalendar()[1], d.month, (d.month-1)//3+1))

REGIONS = ['North','South','East','West']
for i in range(200):
    cur.execute('INSERT INTO dim_customer VALUES (?,?,?)',
                (i+1, random.choice(REGIONS), random.choice(['Retail','Wholesale'])))

CATALOG = [('Phones',299.0),('Laptops',899.0),('Shirts',29.0),('Shoes',79.0),('Kitchen',49.0)]
for i,(cat,price) in enumerate(CATALOG):
    cur.execute('INSERT INTO dim_product VALUES (?,?,?)', (i+1, cat, price))

rows = []
for k in range(60_000):
    d = base + timedelta(days=random.randint(0, 364))
    pk = random.randint(1, 5); qty = random.randint(1, 4)
    rows.append((k+1, d.toordinal(), random.randint(1, 200), pk, qty, qty * CATALOG[pk-1][1]))
cur.executemany('INSERT INTO fact_orders VALUES (?,?,?,?,?,?)', rows)
conn.commit()
print('fact_orders rows:', cur.execute('SELECT COUNT(*) FROM fact_orders').fetchone()[0])
```

```python
@dataclass
class Query:
    name: str
    sql: str
    params: tuple = ()

@dataclass
class ExecPlan:
    query: str
    plan_lines: list
    rows_returned: int
    elapsed_ms: float

def run_and_profile(q: Query) -> ExecPlan:
    plan = [r[3] for r in cur.execute(f'EXPLAIN QUERY PLAN {q.sql}', q.params).fetchall()]
    t0 = time.perf_counter()
    out = cur.execute(q.sql, q.params).fetchall()
    return ExecPlan(q.name, plan, len(out), (time.perf_counter() - t0) * 1000)

SLOW_SQL = '''
SELECT c.region, d.iso_week, SUM(f.net_revenue) AS revenue
  FROM fact_orders f
  JOIN dim_date     d ON f.date_key     = d.date_key
  JOIN dim_customer c ON f.customer_key = c.customer_key
 WHERE d.full_date BETWEEN ? AND ?
 GROUP BY c.region, d.iso_week
 ORDER BY c.region, d.iso_week
'''

baseline = run_and_profile(Query('baseline', SLOW_SQL, ('2024-06-01', '2024-06-30')))
print(f'rows returned : {baseline.rows_returned}')
print(f'elapsed       : {baseline.elapsed_ms:.1f} ms')
print('plan:')
for line in baseline.plan_lines:
    print(' ', line)
```

**Reading the plan.** `SCAN fact_orders` means the engine is walking every row of the largest table and filtering after the join. The `dim_date` filter on `full_date` is a TEXT comparison that can’t use the `date_key` primary key. That’s the smell: the predicate column and the join column are different, so the small-table filter cannot push the join into an index seek on the fact table.

## Part 2 — Index choice (single-column vs composite vs covering)

Indexes are not free: each one slows writes, costs disk, and only helps queries whose predicates match its leading columns. We compare three options against the same query.

```python
def reset_indexes():
    for name, in cur.execute("SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'ix_%'").fetchall():
        cur.execute(f'DROP INDEX {name}')

experiments = []

# A. single column on the join key
reset_indexes()
cur.execute('CREATE INDEX ix_fact_date ON fact_orders(date_key)')
experiments.append(('single-col date_key', run_and_profile(Query('A', SLOW_SQL, ('2024-06-01','2024-06-30')))))

# B. composite that matches GROUP BY's other column
reset_indexes()
cur.execute('CREATE INDEX ix_fact_date_cust ON fact_orders(date_key, customer_key)')
experiments.append(('composite (date_key, customer_key)', run_and_profile(Query('B', SLOW_SQL, ('2024-06-01','2024-06-30')))))

# C. covering index that also carries the measure
reset_indexes()
cur.execute('CREATE INDEX ix_fact_cover ON fact_orders(date_key, customer_key, net_revenue)')
experiments.append(('covering (date_key, customer_key, net_revenue)', run_and_profile(Query('C', SLOW_SQL, ('2024-06-01','2024-06-30')))))

print(f"{'option':50s} {'rows':>6s} {'ms':>8s}  plan summary")
for label, ep in [('baseline (no index)', baseline)] + experiments:
    summary = ' / '.join(p.split(' ')[0] for p in ep.plan_lines)
    print(f'  {label:48s} {ep.rows_returned:>6d} {ep.elapsed_ms:>8.1f}  {summary}')
```

## Part 3 — Decision table: row store vs column store vs OLAP cube

Indexes get you a long way. When the workload is overwhelmingly read-mostly, scan-heavy, and aggregating wide columns, the storage engine itself becomes the lever.

| Engine family | Best when | Avoid when | Examples |
|---|---|---|---|
| **Row store (OLTP)** | High write throughput, point reads, narrow row access | Wide aggregates over millions of rows | Postgres, MySQL InnoDB, SQLite |
| **Columnar (OLAP)** | Scan-heavy aggregates over few columns of a wide table | Heavy point updates, single-row reads | DuckDB, ClickHouse, BigQuery, Snowflake |
| **Pre-aggregated OLAP cube / materialised view** | Same dashboard query runs thousands of times/day | Ad-hoc / unpredictable slice-and-dice | dbt incremental models, Druid, Pinot |
| **Hybrid HTAP** | One workload, both OLTP and OLAP characteristics | Single-purpose workloads where a specialist wins | TiDB, SingleStore, CockroachDB |

Pick by *how the query reads the data*, not by vendor familiarity. Most dashboard latency problems on a row store are solved by either: (a) the right composite index, (b) a nightly materialised view, or (c) moving the dashboard query to a columnar replica. Re-platforming is the last resort.

## Part 4 — Denormalisation tradeoffs

If a join key never changes (the customer’s region for the lifetime of an order) you can fold the dimension column into the fact table. The query stops joining. The cost is duplication and a backfill story for the day someone *does* change a region.

```python
reset_indexes()
cur.execute('ALTER TABLE fact_orders ADD COLUMN region TEXT')
cur.execute('ALTER TABLE fact_orders ADD COLUMN iso_week INTEGER')
cur.execute('''UPDATE fact_orders SET
    region   = (SELECT region   FROM dim_customer c WHERE c.customer_key = fact_orders.customer_key),
    iso_week = (SELECT iso_week FROM dim_date     d WHERE d.date_key     = fact_orders.date_key)''')
cur.execute('CREATE INDEX ix_fact_denorm ON fact_orders(region, iso_week, net_revenue)')
conn.commit()

DENORM_SQL = '''
SELECT region, iso_week, SUM(net_revenue) AS revenue
  FROM fact_orders
 WHERE date_key BETWEEN ? AND ?
 GROUP BY region, iso_week
 ORDER BY region, iso_week
'''
lo = date(2024,6,1).toordinal(); hi = date(2024,6,30).toordinal()
denorm = run_and_profile(Query('denormalised', DENORM_SQL, (lo, hi)))
print(f'denormalised: {denorm.elapsed_ms:.1f} ms over {denorm.rows_returned} rows')
for p in denorm.plan_lines:
    print(' ', p)
```

**When denormalisation is the right answer.**

| Test | Keep normalised | Denormalise |
|---|---|---|
| Dimension value changes per fact row? | Yes — SCD Type 2 in dim | No — safe to copy |
| Reads/writes ratio | < 10:1 | ≫ 100:1 |
| Source of truth lives elsewhere? | Yes — single source | No — fact is canonical |
| Storage cost matters more than CPU? | Yes | No |

If you denormalise, document the invariant ("`fact_orders.region` is the customer’s region *at order time*") and own the backfill job. Silent drift between `dim_customer.region` and `fact_orders.region` is the most common bug in this pattern.

## Part 5 — Regression gate

A tuning win that isn’t protected by CI regresses on the next schema change. Pin a baseline and fail loud when latency — or, equally important, plan shape — drifts.

```python
@dataclass
class PerfBaseline:
    name: str
    max_ms: float
    forbidden_plan_tokens: tuple   # plan must NOT contain these
    required_plan_tokens: tuple    # plan MUST contain these

BASELINES = [
    PerfBaseline('weekly_revenue', max_ms=50.0,
                 forbidden_plan_tokens=('SCAN fact_orders',),
                 required_plan_tokens=('USING INDEX',)),
]

def check(plan: ExecPlan, b: PerfBaseline) -> list:
    issues = []
    if plan.elapsed_ms > b.max_ms:
        issues.append(f'latency {plan.elapsed_ms:.1f}ms > budget {b.max_ms}ms')
    joined = ' | '.join(plan.plan_lines)
    for tok in b.forbidden_plan_tokens:
        if tok in joined:
            issues.append(f'forbidden plan token present: {tok!r}')
    for tok in b.required_plan_tokens:
        if tok not in joined:
            issues.append(f'required plan token missing: {tok!r}')
    return issues

for label, ep in [('baseline', baseline), ('denormalised', denorm)] + experiments:
    issues = check(ep, BASELINES[0])
    verdict = 'PASS' if not issues else 'FAIL'
    print(f'  {label:48s} {ep.elapsed_ms:>7.1f} ms  [{verdict}]')
    for i in issues:
        print(f'       → {i}')
```

**Why gate on plan shape, not just latency.** Latency alone is noisy — cache state, vacuum activity, neighbour load. The plan shape is deterministic given statistics and schema. A passing latency with `SCAN fact_orders` in the plan is a bomb that detonates the moment the table grows. Gate both.

## Part 6 — Runbook: what to try, in what order

| Symptom | First action | If that fails | Last resort |
|---|---|---|---|
| `SCAN <large_table>` in plan | Add composite index leading with the predicate column | Add covering index including the aggregated column | Materialised view refreshed nightly |
| Predicate column ≠ join column | Push filter to the joined dim and propagate via FK | Rewrite predicate to use the indexed surrogate (e.g. `date_key BETWEEN`) | Denormalise the predicate column into the fact |
| GROUP BY explodes intermediate rows | Pre-aggregate in a CTE or subquery | Materialised view at the aggregation grain | Move to columnar engine |
| Plan flips between runs | `ANALYZE` to refresh stats | Pin plan with optimiser hints (where supported) | Stable surrogate index on the join keys |
| Latency spikes after deploy | Diff the plan vs last-good baseline | Roll back if plan shape changed | Pre-deploy plan check in CI |

Work top-down. Each row is roughly 5× cheaper than the next.

## Reflection

**Takeaways**

- The plan is the source of truth. Latency moves; plan shape tells you *why*.
- Indexes are bargains — covering indexes more so — until write rate or storage cost says otherwise.
- Denormalisation is the right tool only when the duplicated value is stable for the fact’s lifetime; otherwise it’s a slow-burning data quality bug.
- The decision between row store, columnar engine, and pre-aggregated cube is set by the *read pattern*, not the team’s favourite vendor.
- A tuning win without a regression gate has a half-life of one sprint.



**Where to go next.** `DS602 ML Project 2` reuses these aggregation patterns to build training tables; `DS801 Distributed Big Data` is what you reach for when a single-node engine — even fully tuned — stops being enough; `DL501 Deep Learning Foundations` then shows why the *quality* of features (not just their availability) governs model behaviour.
