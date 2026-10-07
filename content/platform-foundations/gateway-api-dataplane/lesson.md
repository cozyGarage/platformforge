# Program a Gateway API Data Plane

CRDs let you apply objects. A controller programs them — Envoy Gateway turns Gateway/HTTPRoute into a real data plane on k3d.

## Why it matters

`kubectl get gateway` without Programmed=True is still a sketch. Platform edge migrations need a controller that binds listeners.

## Ticket focus

1. Use a real GatewayClass from Envoy Gateway
2. Apply Gateway; wait for Programmed/Accepted
3. Attach HTTPRoute to payments-api

## Tip codes

- `GW_CLASS_LIVE` — pick an installed GatewayClass (often `eg`)
- `GW_PROGRAMMED` — wait for Programmed=True
- `GW_ROUTE_LIVE` — parentRefs → payments-gw
