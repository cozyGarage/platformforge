# Plan a Gateway API HTTPRoute

Gateway API splits concerns: GatewayClass (controller), Gateway (listeners), HTTPRoute (app routing).

## Why it matters

Classic Ingress mixes cluster and app ownership. Gateway API lets platform teams own Gateways while app teams own HTTPRoutes.

## Ticket focus

1. GatewayClass + Gateway listener for `api.payments.example`
2. HTTPRoute → Service `payments-api:80`
3. Short Ingress → Gateway API migration note

## Tip codes

- `GW_CLASS` — GatewayClass selects the controller
- `GW_LISTENER` — listeners declare hostname/port/protocol
- `HTTP_ROUTE` — parentRefs + backendRefs
- `MIGRATE_INGRESS` — document the Ingress replacement path
