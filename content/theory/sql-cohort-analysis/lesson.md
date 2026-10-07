Group users by acquisition month, track behaviour over time, and measure product health.

```python
import sqlite3, pandas as pd, numpy as np, matplotlib.pyplot as plt
print('Cohort Analysis')
```

## Key Concepts

### What is a Cohort?

A cohort is a group of users sharing a common starting event — most commonly the month they first purchased or signed up (**acquisition cohort**).

### Retention Matrix

$$\text{Retention}_{c,t} = \frac{|\text{users in cohort }c\text{ active in period }t|}{|\text{cohort }c\text{ size}|}$$

Rows = cohorts, columns = periods since acquisition. Period 0 = 100% by definition.

### Key Metrics

**Month-1 retention** — the most actionable early signal. If fewer than 30% of users return after their first month, the product has a core value problem.

**Churn rate at period** $t$: $1 - \text{retention}_t$

**Customer LTV**:

$$\text{LTV} = \text{ARPU} \times \frac{1}{\text{monthly churn rate}}$$

### Reading the Heatmap

- Diagonal unavailability is normal (recent cohorts have less history)
- Horizontal improvement across all cohorts at the same period: product change
- Sudden vertical drop: a specific month had a problem (seasonal, bug)

```python
import sqlite3, pandas as pd, numpy as np, matplotlib.pyplot as plt

rng = np.random.default_rng(42)
conn = sqlite3.connect(":memory:")
conn.execute("CREATE TABLE events (user_id INT, cohort TEXT, period INT)")

rows = []
for ci, cohort in enumerate(["2023-01","2023-02","2023-03","2023-04","2023-05","2023-06"]):
    n = rng.integers(80, 150)
    for u in range(n):
        base_ret = rng.uniform(0.75, 0.92)
        for period in range(8):
            prob = base_ret ** period
            if rng.random() < prob:
                rows.append((ci*150+u, cohort, period))

conn.executemany("INSERT INTO events VALUES (?,?,?)", rows)
conn.commit()

df = pd.read_sql("""
WITH base AS (
    SELECT cohort, COUNT(DISTINCT user_id) AS total FROM events WHERE period=0 GROUP BY cohort
),
act AS (
    SELECT cohort, period, COUNT(DISTINCT user_id) AS active FROM events GROUP BY cohort, period
)
SELECT a.cohort, a.period, ROUND(100.0*a.active/b.total,1) AS retention
FROM act a JOIN base b ON a.cohort=b.cohort
""", conn)

mat = df.pivot(index="cohort", columns="period", values="retention")

fig, axes = plt.subplots(1,2, figsize=(13,5))
im = axes[0].imshow(mat.values, cmap="YlGn", vmin=0, vmax=100, aspect="auto")
axes[0].set_xticks(range(len(mat.columns))); axes[0].set_xticklabels([f"M+{c}" for c in mat.columns])
axes[0].set_yticks(range(len(mat.index))); axes[0].set_yticklabels(mat.index)
for i in range(len(mat.index)):
    for j in range(len(mat.columns)):
        v = mat.values[i,j]
        if not np.isnan(v):
            axes[0].text(j,i,f"{v:.0f}%",ha="center",va="center",fontsize=9,
                        color="white" if v<40 else "black")
plt.colorbar(im, ax=axes[0], label="Retention %")
axes[0].set_title("Cohort Retention Heatmap")

for cohort in mat.index:
    row = mat.loc[cohort].dropna()
    axes[1].plot(row.index, row.values, marker="o", ms=5, lw=2, label=cohort)
axes[1].set_xlabel("Period"); axes[1].set_ylabel("Retention %")
axes[1].set_title("Retention Curves by Cohort"); axes[1].legend(fontsize=8)
axes[1].grid(alpha=0.3); axes[1].spines["top"].set_visible(False); axes[1].spines["right"].set_visible(False)
plt.tight_layout(); plt.show()
conn.close()
```

```python
# Data analysis solution checker

import pandas as pd
import numpy as np

def summarise_df(df, label="DataFrame"):
    """Quick sanity check for a cleaned dataframe."""
    print(f"=== {label} ===")
    print(f"  Shape: {df.shape}")
    print(f"  Missing: {df.isnull().sum().sum()} values ({df.isnull().mean().mean():.1%})")
    print(f"  Dtypes: {df.dtypes.value_counts().to_dict()}")
    if df.select_dtypes('number').shape[1] > 0:
        num = df.select_dtypes('number')
        print(f"  Numeric range: [{num.values.min():.2f}, {num.values.max():.2f}]")
    dupes = df.duplicated().sum()
    print(f"  Duplicates: {dupes}")

def check_clean(df):
    """Assert DataFrame is clean: no NaN, no dupes, correct dtypes."""
    issues = []
    if df.isnull().any().any():
        issues.append(f"{df.isnull().sum().sum()} missing values remain")
    if df.duplicated().any():
        issues.append(f"{df.duplicated().sum()} duplicate rows remain")
    if issues:
        print("✗ ISSUES: " + "; ".join(issues))
    else:
        print("✓ DataFrame is clean")

# Demo
np.random.seed(0)
df = pd.DataFrame({
    'age':    [25, 30, np.nan, 22, 35, 30],
    'salary': [50000, 60000, 75000, 45000, None, 60000],
    'dept':   ['Eng', 'HR', 'Eng', 'Sales', 'HR', 'HR'],
})
summarise_df(df, "Raw")
df_clean = df.dropna().drop_duplicates()
summarise_df(df_clean, "Cleaned")
check_clean(df_clean)
```

```python
# Advanced: Pandas performance patterns

import pandas as pd
import numpy as np
import timeit

np.random.seed(42)
n = 100_000
df = pd.DataFrame({
    'group':   np.random.choice(list('ABCDE'), n),
    'value':   np.random.randn(n),
    'flag':    np.random.randint(0, 2, n).astype(bool),
    'amount':  np.random.exponential(100, n),
})

def time_it(label, fn):
    t = timeit.timeit(fn, number=10) / 10 * 1000
    print(f"  {label:<40}: {t:.2f}ms")

print("Pandas performance patterns:")

# 1. apply vs vectorised operation
time_it("df['value'].apply(lambda x: x**2)", lambda: df['value'].apply(lambda x: x**2))
time_it("df['value'] ** 2  (vectorised)", lambda: df['value'] ** 2)

# 2. iterrows vs vectorised groupby
time_it("groupby().mean()", lambda: df.groupby('group')['amount'].mean())
time_it("groupby().agg({'amount': ['mean', 'std', 'count']})",
        lambda: df.groupby('group')['amount'].agg(['mean', 'std', 'count']))

# 3. Boolean indexing vs query
time_it("df[df.flag & (df.value > 0)]",
        lambda: df[df['flag'] & (df['value'] > 0)])
time_it("df.query('flag and value > 0')",
        lambda: df.query('flag and value > 0'))

print("\nKey insight: always prefer vectorised ops over .apply() or .iterrows()")
```

## Databases & SQL Quick Reference

**Query optimisation checklist:**
1. `EXPLAIN ANALYZE` — find seq scans and high cost nodes
2. Add index on filter/join columns
3. Rewrite correlated subqueries as JOINs or CTEs
4. Use `LIMIT` + keyset pagination instead of `OFFSET`
5. Vacuum/analyse for stale statistics

**Index decision:**
| Situation | Index type |
|-----------|-----------|
| Equality + range | B-Tree (default) |
| Full-text search | GIN with `tsvector` |
| JSONB keys | GIN |
| Geospatial | GiST |
| Sequential insert, range scan | BRIN |

**Transaction isolation levels:**
```sql
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;  -- default Postgres
-- REPEATABLE READ  → no phantom reads for single txn
-- SERIALIZABLE     → full isolation, slowest
```

## Summary & Key Takeaways

- Cohort analysis separates lifecycle effects (decay over time) from calendar effects (product changes, seasonality).
- Period-1 retention is the most actionable early signal — it measures whether users find core value.
- A horizontal line in the heatmap at a late period indicates a loyal retained core — a valuable insight for LTV modelling.
- LTV = ARPU / churn rate; improving early retention compounds over the entire customer lifetime.
- Always check cohort sizes — small cohorts produce noisy curves that look like improvements or regressions.

## 🧠 Quick Quiz

**Q1.** A cohort in cohort analysis is defined by:
- A) Users with the same demographic
- B) **Users who share a common characteristic at a specific time — typically acquisition date** ✓
- C) Users in the same geographic region
- D) Users with similar behaviour

**Q2.** Retention rate in cohort analysis measures:
- A) The proportion of users who paid
- B) **What fraction of a cohort returned in a subsequent time period** ✓
- C) Average revenue per user
- D) Churn rate

**Q3.** The SQL pattern for cohort analysis uses:
- A) GROUP BY only
- B) **Self-join or window functions to compare each user's activity to their cohort's first event** ✓
- C) PIVOT tables
- D) Recursive CTEs

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
