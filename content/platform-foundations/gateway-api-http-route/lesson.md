# Apply a Gateway API HTTPRoute

Gateway API replaces classic Ingress with GatewayClass, Gateway, and HTTPRoute — applied for real on k3d in this lab.

## Why it matters

Sketches do not program a cluster. Applying CRDs objects is how platform teams migrate edge routing with reviewable YAML.

## Ticket focus

1. Apply GatewayClass + Gateway for api.payments.example
2. Apply HTTPRoute to Service payments-api:80
3. Document Ingress → Gateway API migration

## Tip codes

- `GW_CLASS` — GatewayClass selects the controller
- `GW_LISTENER` — listeners declare hostname/port/protocol
- `HTTP_ROUTE` — parentRefs + backendRefs
- `MIGRATE_INGRESS` — document the cutover story
