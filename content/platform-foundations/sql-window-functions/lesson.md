# Rank payments with window functions

Window functions answer reporting questions without collapsing rows. They let you rank events and calculate running totals while keeping each payment visible.

## Schema

`payments(id, account, amount, created_at)`

Seeded accounts:

- alice: 25 then 40
- bob: 80 then 20
- carol: 120

## Tasks

1. Write `/workspace/sql/window.sql`
2. Use `ROW_NUMBER() OVER (PARTITION BY account ORDER BY created_at)`
3. Use `SUM(amount) OVER (PARTITION BY account ORDER BY created_at)`
4. Save results to `/workspace/sql/window.txt`
5. Confirm Alice reaches running total `65` and Bob reaches `100`
