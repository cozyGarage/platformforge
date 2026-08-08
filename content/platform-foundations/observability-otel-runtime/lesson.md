# Run an OpenTelemetry Collector on k3d

Planning pipelines is step one. This lab runs Jaeger + an OTel Collector on k3d and makes you prove the traces pipeline exports.

## Why it matters

A ConfigMap that never rolls out is documentation. Ready Deployments with a Jaeger exporter are a platform contract.

## Ticket focus

1. Prove jaeger + otel-collector Available
2. Apply collector config that exports to Jaeger
3. Document Collector / Jaeger runtime notes

## Tip codes

- `OTEL_READY` — both Deployments Available in observability
- `OTEL_EXPORT_LIVE` — otlp/jaeger exporter to jaeger.observability.svc
