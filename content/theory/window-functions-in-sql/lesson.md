Row-level analytics: ranking, running totals, and lag/lead comparisons without collapsing rows.

```python
import sqlite3, pandas as pd, numpy as np, matplotlib.pyplot as plt
conn = sqlite3.connect(':memory:')
print('Window Functions')
```

## Key Concepts

### Syntax

```sql
function_name() OVER (
    [PARTITION BY column, ...]
    [ORDER BY column]
    [ROWS BETWEEN ... AND ...]
)
```

**PARTITION BY** resets the window per group (like GROUP BY but keeps all rows).  
**ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW** = running total.

### Ranking Functions

| Function | Behaviour |
|---|---|
| `ROW_NUMBER()` | Unique 1,2,3 even for ties |
| `RANK()` | Ties share rank; skips next (1,1,3) |
| `DENSE_RANK()` | Ties share rank; no skipping (1,1,2) |
| `NTILE(n)` | Divide into n equal buckets |

### Offset Functions

- `LAG(col, n)` — value n rows before current row
- `LEAD(col, n)` — value n rows after current row
- `FIRST_VALUE(col)`, `LAST_VALUE(col)` — first/last in the frame

### Aggregate Window Functions

Any aggregate (`SUM`, `AVG`, `MIN`, `MAX`) works as a window function:

```sql
SUM(revenue) OVER (
    PARTITION BY region
    ORDER BY month
    ROWS UNBOUNDED PRECEDING
) AS running_total
```

### GROUP BY vs OVER

`GROUP BY` collapses rows; `OVER` adds a column while keeping all rows.

```python
import sqlite3, pandas as pd, numpy as np, matplotlib.pyplot as plt

conn = sqlite3.connect(":memory:")
conn.executescript("""
CREATE TABLE sales (month TEXT, region TEXT, rep TEXT, revenue REAL);
INSERT INTO sales VALUES
('2024-01','North','Alice',12000),('2024-01','North','Bob',9500),
('2024-01','South','Carol',15000),('2024-01','South','Dave',11000),
('2024-02','North','Alice',13500),('2024-02','North','Bob',9500),
('2024-02','South','Carol',14000),('2024-02','South','Dave',12500),
('2024-03','North','Alice',11000),('2024-03','North','Bob',10500),
('2024-03','South','Carol',16000),('2024-03','South','Dave',13000),
('2024-04','North','Alice',14500),('2024-04','North','Bob',8500),
('2024-04','South','Carol',15500),('2024-04','South','Dave',14000);
""")

rank_df = pd.read_sql("""
SELECT month, region, rep, revenue,
       RANK()       OVER (PARTITION BY month ORDER BY revenue DESC) AS rnk,
       ROW_NUMBER() OVER (PARTITION BY month ORDER BY revenue DESC) AS row_num
FROM sales ORDER BY month, rnk
""", conn)
print(rank_df.head(8).to_string(index=False))

running_df = pd.read_sql("""
SELECT month, region,
       SUM(revenue) AS monthly_rev,
       SUM(SUM(revenue)) OVER (PARTITION BY region ORDER BY month
                               ROWS UNBOUNDED PRECEDING) AS running_total,
       LAG(SUM(revenue)) OVER (PARTITION BY region ORDER BY month) AS prev_month
FROM sales GROUP BY month, region ORDER BY region, month
""", conn)
print(running_df.to_string(index=False))

fig, axes = plt.subplots(1,2, figsize=(12,4))
for region, grp in running_df.groupby("region"):
    axes[0].plot(grp["month"], grp["running_total"]/1000, marker="o", label=region, lw=2)
axes[0].set_title("Cumulative Revenue by Region"); axes[0].set_ylabel("$k")
axes[0].legend(); axes[0].grid(alpha=0.3); axes[0].tick_params(axis="x",rotation=30)

top1 = rank_df[rank_df["rnk"]==1].groupby("rep").size() / rank_df["month"].nunique()
axes[1].bar(top1.index, top1.values*100, color="#3498db", alpha=0.8)
axes[1].set_title("% of Months as Top Earner"); axes[1].set_ylabel("%")
axes[1].grid(axis="y", alpha=0.3); axes[1].spines["top"].set_visible(False); axes[1].spines["right"].set_visible(False)
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

- Window functions add a computed column without reducing row count — unlike GROUP BY.
- PARTITION BY resets the window per group; omitting it uses the entire table as one window.
- LAG/LEAD enable period-over-period comparisons in a single query — no self-join needed.
- Running totals use `SUM(...) OVER (... ROWS UNBOUNDED PRECEDING)`.
- Choose ROW_NUMBER for unique rankings, RANK when ties should share a rank with gaps, DENSE_RANK with no gaps.

## 🧠 Quick Quiz

**Q1.** Window functions differ from GROUP BY because:
- A) They are faster
- B) They only work with aggregate functions
- C) **They do not collapse rows — each row keeps its own result** ✓
- D) They require an ORDER BY clause

**Q2.** `ROW_NUMBER() OVER (PARTITION BY dept ORDER BY salary DESC)` gives:
- A) Salary rank across all employees
- B) **Row rank within each department, ordered by salary descending** ✓
- C) Running total of salaries per department
- D) Average salary per department

**Q3.** `LAG(value, 1)` returns:
- A) The next row's value
- B) **The previous row's value** ✓
- C) The first row's value in the partition
- D) The running minimum

<details><summary>Answers</summary>1-C, 2-B, 3-B</details>
