# Plan an OpenTelemetry Collector Pipeline

Apps emit OTLP. The collector is where you shape, sample, and route telemetry before backends see it.

## Why it matters

Pointing every service at Tempo and Prometheus directly creates fan-out and inconsistent sampling. A collector pipeline is the platform contract.

## Ticket focus

1. Six-line pipeline: OTLP → processors → Tempo/Prometheus
2. Ops notes for receivers, processors, exporters

## Tip codes

- `OTEL_PIPELINE` — receivers → processors → exporters in service.pipelines
- `OTEL_RECV` — OTLP as the single ingest contract
- `OTEL_EXPORT` — traces to Tempo/Jaeger, metrics to Prometheus
