# Join accounts and payments

Normalized schemas need joins. Platform on-call work often means stitching identity tables to event tables.

## Schema

- `accounts(id, name)`
- `payments(id, account_id, amount)`

## Tasks

1. Write `/workspace/sql/by_account.sql` joining both tables
2. Return each account `name` with `SUM(amount)`
3. Save output to `/workspace/sql/by_account.txt` (alice=65, carol=120)
