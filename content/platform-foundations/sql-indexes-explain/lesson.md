# Speed queries with indexes

Without an index, every account filter is a table scan. `EXPLAIN QUERY PLAN` shows whether SQLite can use your index.

## Tasks

1. Create `/workspace/sql/index.sql` for `idx_payments_account` on `payments(account)` and apply it
2. Write `/workspace/sql/lookup.sql` filtering `account = 'alice'`
3. Save `EXPLAIN QUERY PLAN` output to `/workspace/sql/explain.txt` (must mention the index)
