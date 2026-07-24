# Restrict pod traffic with NetworkPolicy

CKA expects you to isolate east-west traffic. A common pattern is default-deny Ingress on sensitive pods, then allow only known clients.

## Starting state

Namespace `payments` has Deployments `api` (`app=api`) and `frontend` (`app=frontend`).

## Tasks

1. Write `/workspace/networkpolicy.yaml`
2. Select pods `app=api`
3. Default-deny Ingress, then allow from `app=frontend` on TCP/80
4. `kubectl -n payments apply -f /workspace/networkpolicy.yaml`

## Example

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-ingress
  namespace: payments
spec:
  podSelector:
    matchLabels:
      app: api
  policyTypes: [Ingress]
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: frontend
      ports:
        - protocol: TCP
          port: 80
```
