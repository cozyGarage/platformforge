# Query payments with SQL

Platform engineers read databases during incidents. This lab uses SQLite as a stand-in ledger.

## Schema

`payments(id, account, amount)` with rows alice/25, bob/80, carol/120.

## Tasks

1. `/workspace/sql/total.sql` — sum all amounts → result `225` in `total.txt`
2. `/workspace/sql/high_value.sql` — accounts with `amount >= 100` → `high_value.txt` contains `carol`
