Period-over-period comparisons, rolling averages, and gap filling.

```python
import sqlite3, pandas as pd, numpy as np, matplotlib.pyplot as plt
print('Time-Series SQL')
```

## Key Concepts

### Date Truncation

```sql
strftime('%Y-%m', date)  -- SQLite: truncate to month
strftime('%Y-W%W', date) -- SQLite: ISO week
DATE_TRUNC('month', ts)  -- PostgreSQL/BigQuery
```

### Period-over-Period Comparison

```sql
SELECT month, revenue,
       LAG(revenue,  1) OVER (ORDER BY month) AS prev_month,
       LAG(revenue, 12) OVER (ORDER BY month) AS prev_year,
       ROUND(100.0*(revenue - LAG(revenue,12) OVER(ORDER BY month))
             / LAG(revenue,12) OVER(ORDER BY month),1) AS yoy_pct
FROM monthly_revenue
```

### Rolling Average

```sql
AVG(revenue) OVER (
    ORDER BY month
    ROWS BETWEEN 2 PRECEDING AND CURRENT ROW
) AS rolling_3m
```

### Gap Filling

SQL only returns rows that exist. Fill missing dates with a calendar CTE + LEFT JOIN:

```sql
WITH calendar(date) AS (
    SELECT '2024-01-01'
    UNION ALL SELECT date(date, '+1 day') FROM calendar WHERE date < '2024-12-31'
)
SELECT c.date, COALESCE(d.revenue, 0)
FROM calendar c LEFT JOIN daily_rev d ON c.date = d.date
```

### Time Zone Best Practice

Store all timestamps in **UTC**; convert to local time only in the final SELECT display.

```python
import sqlite3, pandas as pd, numpy as np, matplotlib.pyplot as plt

rng = np.random.default_rng(42)
conn = sqlite3.connect(":memory:")
dates = pd.date_range("2022-01-01","2024-03-31",freq="D")
trend = np.linspace(1000,1800,len(dates))
seasonal = 200*np.sin(2*np.pi*np.arange(len(dates))/365.25)
rev = trend + seasonal + rng.normal(0,80,len(dates))
conn.execute("CREATE TABLE daily_rev (date TEXT, revenue REAL)")
conn.executemany("INSERT INTO daily_rev VALUES (?,?)",
                 [(d.strftime("%Y-%m-%d"),max(0,r)) for d,r in zip(dates,rev)])
conn.commit()

analysis = pd.read_sql("""
WITH monthly AS (
    SELECT strftime('%Y-%m', date) AS month, SUM(revenue) AS rev
    FROM daily_rev GROUP BY month
)
SELECT month, rev,
       LAG(rev,12) OVER (ORDER BY month) AS prev_year,
       AVG(rev)    OVER (ORDER BY month ROWS BETWEEN 2 PRECEDING AND CURRENT ROW) AS rolling_3m,
       ROUND(100.0*(rev - LAG(rev,12) OVER(ORDER BY month))
             / LAG(rev,12) OVER(ORDER BY month),1) AS yoy_pct
FROM monthly
""", conn)

fig, axes = plt.subplots(2,2,figsize=(13,8))
fig.suptitle("Time-Based SQL Analysis",fontsize=13)

ax=axes[0,0]
ax.plot(analysis["month"], analysis["rev"]/1000, alpha=0.4, color="#aaaaaa", label="Monthly")
ax.plot(analysis["month"], analysis["rolling_3m"]/1000, color="#3498db", lw=2.5, label="3m rolling")
ax.set_title("Monthly Revenue + 3m Rolling Avg"); ax.set_ylabel("$k"); ax.legend()
ax.grid(alpha=0.2); ax.tick_params(axis="x",rotation=45); ax.set_xticks(ax.get_xticks()[::4])

yoy = analysis.dropna(subset=["yoy_pct"])
ax=axes[0,1]
colors = ["#2ecc71" if v>=0 else "#e74c3c" for v in yoy["yoy_pct"]]
ax.bar(yoy["month"], yoy["yoy_pct"], color=colors, alpha=0.8)
ax.axhline(0,color="black",lw=0.8)
ax.set_title("Year-over-Year Growth %"); ax.set_ylabel("YoY %")
ax.grid(axis="y",alpha=0.2); ax.tick_params(axis="x",rotation=45); ax.set_xticks(ax.get_xticks()[::3])

gap = pd.read_sql("""
WITH calendar(date) AS (
    SELECT '2024-01-01'
    UNION ALL SELECT date(date,'+1 day') FROM calendar WHERE date < '2024-01-31'
)
SELECT c.date, COALESCE(d.revenue,0) AS revenue
FROM calendar c LEFT JOIN daily_rev d ON c.date=d.date ORDER BY c.date
""", conn)
ax=axes[1,0]
ax.fill_between(gap["date"], gap["revenue"]/1000, alpha=0.4, color="steelblue")
ax.plot(gap["date"], gap["revenue"]/1000, color="steelblue", lw=1.5)
ax.set_title("Gap Filling: Calendar LEFT JOIN (Jan 2024)")
ax.set_ylabel("$k"); ax.tick_params(axis="x",rotation=45); ax.set_xticks(ax.get_xticks()[::5])
ax.grid(alpha=0.2)

mom_pct = analysis["rev"].pct_change()*100
ax=axes[1,1]
ax.hist(mom_pct.dropna(), bins=20, color="#9b59b6", alpha=0.8, edgecolor="white")
ax.axvline(mom_pct.mean(), color="red", ls="--", lw=2, label=f"Mean: {mom_pct.mean():.1f}%")
ax.set_title("MoM Growth Distribution"); ax.set_xlabel("Growth %"); ax.legend(); ax.grid(alpha=0.2)
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

## SQL Quick Reference

```sql
-- Window functions
SELECT name, salary,
  RANK() OVER (PARTITION BY dept ORDER BY salary DESC) AS rank,
  AVG(salary) OVER (PARTITION BY dept) AS dept_avg
FROM employees;

-- CTEs
WITH ranked AS (
  SELECT *, ROW_NUMBER() OVER (ORDER BY created_at) AS rn FROM orders
)
SELECT * FROM ranked WHERE rn <= 10;
```

**Index types:**
| Type | Best for |
|------|----------|
| B-Tree | Equality + range queries |
| Hash | Exact equality only |
| GIN | Full-text, array, JSONB |
| BRIN | Naturally ordered large tables |

**Query plan reading:** `EXPLAIN ANALYZE SELECT ...`

## Summary & Key Takeaways

- strftime (SQLite) and DATE_TRUNC (PostgreSQL) bucket timestamps into time periods; always be explicit about granularity.
- YoY comparison (LAG 12 months) removes seasonal effects; MoM is noisier but more timely.
- Rolling averages require an explicit ROWS BETWEEN frame — without it, behaviour is undefined or database-specific.
- Gap filling requires a calendar CTE + LEFT JOIN; omitting it hides zero-revenue periods from trend analysis.
- Always store timestamps in UTC and convert to local time only at display layer.

## 🧠 Quick Quiz

**Q1.** A date spine (or calendar table) is used in time series SQL to:
- A) Store timezone conversions
- B) **Ensure every date appears in results even when no data exists for that date** ✓
- C) Partition tables by month
- D) Calculate date differences

**Q2.** Rolling 7-day average in SQL uses:
- A) GROUP BY with 7 rows
- B) **`AVG() OVER (ORDER BY date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW)`** ✓
- C) A subquery per date
- D) CASE WHEN with date arithmetic

**Q3.** TimescaleDB extends PostgreSQL for time series by:
- A) Adding OLAP features
- B) **Automatically partitioning time-series tables into time-based "chunks"** ✓
- C) Providing a columnar storage engine
- D) Adding streaming ingestion

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
