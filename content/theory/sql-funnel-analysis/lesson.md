Track users through ordered steps, identify bottlenecks, and A/B test improvements.

```python
import sqlite3, pandas as pd, numpy as np, matplotlib.pyplot as plt
print('Funnel Analysis')
```

## Key Concepts

### Funnel Definition

A **conversion funnel** is an ordered sequence of steps. Each step has:
- **Entry count** — users who reached this step
- **Completion count** — users who completed it
- **Step CVR** = completions / entries
- **Overall CVR** = final completions / step-1 entries

### SQL Pattern

```sql
SELECT
    COUNT(DISTINCT CASE WHEN step_num >= 1 THEN user_id END) AS step1,
    COUNT(DISTINCT CASE WHEN step_num >= 2 THEN user_id END) AS step2,
    COUNT(DISTINCT CASE WHEN step_num >= 3 THEN user_id END) AS step3
FROM funnel_events
```

### A/B Split in Funnels

$$\text{lift} = \frac{\text{CVR}_B - \text{CVR}_A}{\text{CVR}_A}$$

Use **Fisher's exact test** for statistical significance on count data.

### Prioritising Drop-off

Focus on the step with the largest **absolute** user loss (1,000 lost at 80% > 50 lost at 20%).

```python
import sqlite3, pandas as pd, numpy as np, matplotlib.pyplot as plt
from scipy.stats import fisher_exact

rng = np.random.default_rng(42)
conn = sqlite3.connect(":memory:")

steps = ["Landing","Product View","Add to Cart","Checkout","Purchase"]
rates_a = [1.0, 0.60, 0.40, 0.70, 0.65]
rates_b = [1.0, 0.62, 0.45, 0.72, 0.70]
n = 1000

rows = []
for uid in range(n*2):
    var = "A" if uid < n else "B"
    rates = rates_a if var=="A" else rates_b
    active = True
    for i,(step,p) in enumerate(zip(steps,rates)):
        if not active: break
        if rng.random() < p:
            rows.append((uid, var, step, i+1))
        else:
            active = False

conn.execute("CREATE TABLE funnel (user_id INT, variant TEXT, step TEXT, step_num INT)")
conn.executemany("INSERT INTO funnel VALUES (?,?,?,?)", rows)
conn.commit()

df = pd.read_sql("""
SELECT variant, step_num, step, COUNT(DISTINCT user_id) AS users
FROM funnel GROUP BY variant, step_num, step ORDER BY variant, step_num
""", conn)

fig, axes = plt.subplots(1,3,figsize=(15,5))
for vi, var in enumerate(["A","B"]):
    grp = df[df["variant"]==var].sort_values("step_num")
    total = grp["users"].iloc[0]
    grp = grp.copy(); grp["pct"] = grp["users"]/total*100
    col = "#3498db" if var=="A" else "#e74c3c"
    bars = axes[vi].barh(grp["step"][::-1], grp["pct"][::-1], color=col, alpha=0.75)
    for bar, (_, row) in zip(bars, grp[::-1].iterrows()):
        axes[vi].text(bar.get_width()+1, bar.get_y()+bar.get_height()/2,
                f"{row['pct']:.1f}% ({int(row['users'])})", va="center", fontsize=9)
    axes[vi].set_xlim(0,115); axes[vi].set_title(f"Variant {var} Funnel")
    axes[vi].set_xlabel("% of entrants")
    axes[vi].spines["top"].set_visible(False); axes[vi].spines["right"].set_visible(False)

cmp = df.pivot_table(index=["step","step_num"],columns="variant",values="users").reset_index().sort_values("step_num")
x = np.arange(len(cmp)); w=0.35; ta=n; tb=n
axes[2].bar(x-w/2, cmp["A"]/ta*100, w, label="A", color="#3498db", alpha=0.8)
axes[2].bar(x+w/2, cmp["B"]/tb*100, w, label="B", color="#e74c3c", alpha=0.8)
axes[2].set_xticks(x); axes[2].set_xticklabels(cmp["step"], rotation=25, ha="right", fontsize=8)
axes[2].set_title("A/B Comparison"); axes[2].legend(); axes[2].grid(axis="y",alpha=0.3)
axes[2].spines["top"].set_visible(False); axes[2].spines["right"].set_visible(False)
plt.tight_layout(); plt.show()

a_conv = len([r for r in rows if r[1]=="A" and r[3]==5])
b_conv = len([r for r in rows if r[1]=="B" and r[3]==5])
_, pval = fisher_exact([[a_conv,n-a_conv],[b_conv,n-b_conv]])
print(f"Final CVR  A:{a_conv/n:.3f}  B:{b_conv/n:.3f}  p={pval:.4f} {'*significant*' if pval<0.05 else 'not significant'}")
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

- Focus optimisation on the step with the largest **absolute** user drop, not the lowest conversion rate.
- Step conversion = completions / entrants at that step; overall conversion = final completions / step-1 entrants.
- A/B test funnel variants with Fisher's exact test or chi-squared for statistical significance.
- Session-based funnels (strict ordering within a session) are more precise than looser multi-session funnels.
- Segment funnels by device, channel, and user type — the same bottleneck often has different root causes across segments.

## 🧠 Quick Quiz

**Q1.** A conversion funnel tracks:
- A) Revenue per user
- B) **The percentage of users completing each step in a defined sequence of actions** ✓
- C) Session duration
- D) Page views

**Q2.** Drop-off rate at funnel step n is:
- A) Users at step n divided by total users
- B) **Users who completed step n-1 but not step n, as a fraction of step n-1** ✓
- C) Bounce rate at step n
- D) Time spent at step n

**Q3.** Ordered funnel analysis in SQL requires:
- A) PIVOT operations
- B) **Window functions with LEAD/LAG to verify event sequence and ordering** ✓
- C) Recursive CTEs
- D) CASE WHEN aggregations only

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
