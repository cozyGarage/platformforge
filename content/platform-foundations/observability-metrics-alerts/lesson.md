# Turn metrics into an alert

Logs tell stories; metrics tell rates. On-call engineers alert on sustained error rates, not single log lines — Boot.dev **Learn Logging and Observability** connects the two.

## Starting state

`/workspace/metrics/api.prom` has Prometheus-style counters:

```
http_requests_total{status="200"} 900
http_requests_total{status="500"} 100
```

## Tasks

1. Compute error rate percent: `100 * 500s / (200s + 500s)` → write `10` to `/workspace/error-rate.txt`
2. Author `/workspace/alerts/api-high-errors.yaml` with alert name `APIHighErrorRate` and threshold `5`
3. Write `/workspace/runbook.txt` with a one-line note that mentions `rollback`

## Example alert

```yaml
alert: APIHighErrorRate
expr: error_rate_percent > 5
for: 5m
labels:
  severity: page
```

## Example commands

```sh
echo 10 > /workspace/error-rate.txt
printf 'alert: APIHighErrorRate\nexpr: error_rate_percent > 5\n' > /workspace/alerts/api-high-errors.yaml
echo 'If APIHighErrorRate fires, rollback the last deploy and check /pay handlers.' > /workspace/runbook.txt
```
