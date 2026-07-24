# Point-in-time recovery drill

PostgreSQL PITR restores a base backup, then replays write-ahead log (WAL) segments up to a chosen time — stopping before a bad migration or delete. This lab simulates that workflow with CSV + SQL segments so you practice the recovery decision without a full database daemon.

## Incident

At `12:15` someone ran `DELETE FROM accounts`. Live data in `/workspace/data/accounts.csv` is corrupted. Your recovery target is earlier:

```
/workspace/recovery.target → 2026-07-24T12:12:00Z
```

## Artifacts

| Path | Role |
| --- | --- |
| `backup/base.csv` | Base backup at 12:00 |
| `wal/000001.sql` | Alice balance → 90 at 12:05 |
| `wal/000002.sql` | Insert carol at 12:10 |
| `wal/000003.sql` | Wipe at 12:15 — **do not apply** |

## Tasks

1. Read `recovery.target` and only apply WAL whose timestamps are `<=` the target
2. Reconstruct the table state into `/workspace/data/accounts.csv`
3. Confirm alice=`90`, bob=`50`, carol=`25`

## Expected recovered CSV

```csv
id,account,balance
1,alice,90
2,bob,50
3,carol,25
```
