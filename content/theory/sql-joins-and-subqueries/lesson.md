## SQL joins subqueries filtering (Vault Research)

**Insight 1:**
> Byte Pair Encoding (BPE) tokenizers eliminate the out-of-vocabulary problem by recursively merging the most frequent character pairs in a corpus until

Source: *Build a Large Language Model (From Scratch)* (Chunk 3167)

**Insight 2:**
> The size-overlap trade-off has a sweet spot: small chunks (200–500 chars) improve precision but lose context; large chunks (1k–2k) carry context but d

Source: *30 Agents Every AI Engineer Must Build - 1 Edition - Transform LLMs into autonomous decision-making vertical agents in healthcare, finance, and beyond - 30 Agents Every AI Engineer Must Build (for )* (Chunk 3629)

**Insight 3:**
> In LangGraph supervisor architectures, all subagent messages are typically appended to a shared message list so subagents see each other's work; alter

Source: *Learning LangChain* (Chunk 6950)

```python
import sqlite3

conn = sqlite3.connect(':memory:')
c = conn.cursor()
c.executescript('''
CREATE TABLE customers (id INTEGER PRIMARY KEY, name TEXT, city TEXT);
CREATE TABLE orders (id INTEGER PRIMARY KEY, customer_id INTEGER, amount REAL, date TEXT);
INSERT INTO customers VALUES (1,'Alice','NYC'),(2,'Bob','LA'),(3,'Charlie','NYC'),(4,'Diana','Chicago');
INSERT INTO orders VALUES (1,1,150,'2024-01-15'),(2,1,200,'2024-02-20'),
(3,2,75,'2024-01-10'),(4,3,300,'2024-03-05'),(5,1,125,'2024-03-15');
''')

join_queries = [
    ('INNER JOIN', '''SELECT c.name, o.amount, o.date
FROM customers c INNER JOIN orders o ON c.id = o.customer_id'''),
    ('LEFT JOIN (includes Diana with no orders)',
     '''SELECT c.name, COALESCE(SUM(o.amount), 0) as total
FROM customers c LEFT JOIN orders o ON c.id = o.customer_id
GROUP BY c.name'''),
    ('Subquery: customers with above-avg orders',
     '''SELECT name FROM customers WHERE id IN
(SELECT customer_id FROM orders WHERE amount > (SELECT AVG(amount) FROM orders))'''),
    ('CTE: customer order summary',
     '''WITH order_summary AS (
    SELECT customer_id, COUNT(*) as n_orders, SUM(amount) as total
    FROM orders GROUP BY customer_id
)
SELECT c.name, os.n_orders, os.total
FROM customers c JOIN order_summary os ON c.id = os.customer_id
ORDER BY os.total DESC'''),
]

for title, sql in join_queries:
    print(f'\n--- {title} ---')
    for row in c.execute(sql).fetchall():
        print(f'  {row}')
conn.close()
```

## Exercises

### Exercise 1 (⭐): Practice the queries from this chapter.
### Exercise 2 (⭐⭐): Write more complex queries.
---
*Next chapter*

## JOIN Types

```sql
INNER JOIN  -- only matching rows from both tables
LEFT JOIN   -- all from left, NULLs where no right match
RIGHT JOIN  -- all from right, NULLs where no left match
FULL JOIN   -- all rows from both, NULLs for non-matches
CROSS JOIN  -- cartesian product (every combination)
SELF JOIN   -- join table to itself (hierarchy, adjacency)
```

```python
import sqlite3

conn = sqlite3.connect(':memory:')
conn.executescript("""
    CREATE TABLE emp (id INT, name TEXT, mgr_id INT, dept_id INT);
    CREATE TABLE dept (id INT, name TEXT);
    INSERT INTO dept VALUES (1,'Eng'),(2,'Mkt'),(3,'HR');
    INSERT INTO emp VALUES
        (1,'Alice',NULL,1),(2,'Bob',1,1),(3,'Charlie',1,2),
        (4,'Diana',NULL,3),(5,'Eve',4,NULL);  -- Eve has no dept
""")

# INNER JOIN: only employees with a department
print('INNER JOIN (employees with depts):')
for r in conn.execute('SELECT e.name, d.name FROM emp e INNER JOIN dept d ON e.dept_id = d.id'):
    print(f'  {r[0]:10s} -> {r[1]}')

# LEFT JOIN: all employees, NULL if no dept
print('\nLEFT JOIN (all employees):')
for r in conn.execute('SELECT e.name, d.name FROM emp e LEFT JOIN dept d ON e.dept_id = d.id'):
    print(f'  {r[0]:10s} -> {r[1]}')

# SELF JOIN: employee with their manager
print('\nSELF JOIN (employee + manager):')
for r in conn.execute('SELECT e.name, m.name as manager FROM emp e LEFT JOIN emp m ON e.mgr_id = m.id'):
    print(f'  {r[0]:10s} manager: {r[1]}')
```

```python
# Subqueries: scalar, correlated, and EXISTS

conn2 = sqlite3.connect(':memory:')
conn2.executescript("""
    CREATE TABLE orders (id INT, customer_id INT, amount REAL, date TEXT);
    INSERT INTO orders VALUES
        (1,100,50.0,'2025-01'),(2,100,120.0,'2025-02'),
        (3,200,30.0,'2025-01'),(4,300,200.0,'2025-02'),
        (5,200,80.0,'2025-02');
""")

# Scalar subquery: orders above average
print('Orders above average amount:')
for r in conn2.execute('SELECT id, amount FROM orders WHERE amount > (SELECT AVG(amount) FROM orders)'):
    print(f'  Order {r[0]}: ${r[1]}')

# Correlated subquery: customers with more than 1 order
print('\nCustomers with >1 order (correlated subquery):')
q = '''
    SELECT DISTINCT customer_id
    FROM orders o1
    WHERE (SELECT COUNT(*) FROM orders o2 WHERE o2.customer_id = o1.customer_id) > 1
'''
for r in conn2.execute(q):
    print(f'  customer_id: {r[0]}')

# IN subquery vs JOIN (performance note)
print('\nIN subquery vs JOIN:')
q_in = 'SELECT id FROM orders WHERE customer_id IN (SELECT customer_id FROM orders WHERE date = "2025-01")'
q_join = 'SELECT DISTINCT o2.id FROM orders o1 JOIN orders o2 ON o1.customer_id = o2.customer_id WHERE o1.date = "2025-01"'
in_results = [r[0] for r in conn2.execute(q_in)]
join_results = sorted(set(r[0] for r in conn2.execute(q_join)))
print(f'  IN result:   {sorted(in_results)}')
print(f'  JOIN result: {join_results}')
```

## Further Reading

**Books:**
- *Python for Data Analysis* (McKinney) — pandas deep dive
- *Storytelling with Data* (Nussbaumer Knaflic) — visualisation
- *Designing Data-Intensive Applications* (Kleppmann) — systems

**Resources:**
- [Pandas docs](https://pandas.pydata.org/docs/) — API reference
- [SQLZoo](https://sqlzoo.net) / [Mode Analytics SQL Tutorial](https://mode.com/sql-tutorial/)
- [Kaggle Learn](https://kaggle.com/learn) — free data science courses

## Performance: JOINs and Indexes

| Operation | Without Index | With Index |
|-----------|--------------|------------|
| JOIN on foreign key | O(N×M) | O(N log M) |
| WHERE on column | O(N) | O(log N) |
| ORDER BY | O(N log N) | O(1) if indexed |

```sql
CREATE INDEX idx_emp_dept ON employees(dept_id);
CREATE INDEX idx_orders_cust_date ON orders(customer_id, order_date);
```

**EXPLAIN QUERY PLAN** in SQLite or **EXPLAIN ANALYZE** in PostgreSQL shows if indexes are used.

## Key Takeaways

- **INNER JOIN**: Returns only rows where the join condition matches in both tables, discarding non-matching rows from either side.
- **LEFT/RIGHT/FULL OUTER JOIN**: Preserves all rows from the left, right, or both tables respectively, filling NULLs where no match exists.
- **Correlated Subqueries**: Subqueries that reference the outer query's columns and re-execute for each outer row — powerful but can be slow.
- **EXISTS vs IN**: EXISTS short-circuits on first match and handles NULLs correctly; IN can be faster for small, static value sets.
- **Index Utilization on JOINs**: Creating indexes on join columns (foreign keys) can reduce nested-loop join complexity from O(n²) to O(n log n).

## 🧠 Quick Quiz

**Q1.** An INNER JOIN returns:
- A) All rows from the left table
- B) All rows from both tables
- C) **Only rows with matching keys in both tables** ✓
- D) All rows from the right table

**Q2.** A LEFT JOIN includes:
- A) Only matched rows
- B) **All rows from the left table, with NULLs for non-matching right rows** ✓
- C) All rows from the right table with NULLs for non-matching left rows
- D) The Cartesian product

**Q3.** A correlated subquery differs from a regular subquery because:
- A) It is faster
- B) It returns multiple rows
- C) **It references the outer query and executes once per outer row** ✓
- D) It uses aggregate functions

<details><summary>Answers</summary>1-C, 2-B, 3-C</details>
