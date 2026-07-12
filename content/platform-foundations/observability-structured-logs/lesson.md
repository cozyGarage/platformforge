# Parse structured logs

Structured JSON logs power modern observability stacks. Boot.dev **Learn Logging and Observability** starts here — filtering events before they hit Loki or CloudWatch.

## Data

`/workspace/logs/api.jsonl` contains JSON lines from an API gateway.

## Tasks

1. Extract lines where `status` is `500` into `/workspace/errors.jsonl`
2. Write the count of 500 errors to `/workspace/error-count.txt`

## Example

```sh
jq -c 'select(.status == 500)' /workspace/logs/api.jsonl > /workspace/errors.jsonl
wc -l < /workspace/errors.jsonl > /workspace/error-count.txt
```
