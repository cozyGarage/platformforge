# Capture database changes with CDC

Change Data Capture streams inserts/updates/deletes from Postgres into Kafka (often via Debezium). Platform teams use CDC for ledgers, search indexes, and audit sinks — without batch ETL.

## Starting state

- `connector/payments.json` is missing `topic.prefix`
- `events/payments.jsonl` has create, update, and delete events
- `sink/` is empty

## Tasks

1. Set `topic.prefix` to `fintech` in the connector config
2. Materialize create/update rows into `/workspace/sink/payments.csv` as `id,amount,op` (skip deletes)
3. Write total raw event count (`3`) to `/workspace/sink/event-count.txt`

## Expected sink (example)

```csv
id,amount,op
1,40,u
```

(or include the create row too if you keep history — validators require the final `1,40` state and no delete row for id 2)
