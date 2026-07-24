# Route external traffic with Ingress

Services give you a stable ClusterIP. Ingress (or Gateway API) maps hostnames and paths to those Services at the edge — a core CKA skill.

## Starting state

Deployment + Service `api` already exist on port 80.

## Tasks

1. Write `/workspace/ingress.yaml` for Ingress `api`
2. Host `api.payments.local`, path `/` (Prefix) → Service `api:80`
3. Apply and verify with `kubectl get ingress api`

## Example

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api
spec:
  rules:
    - host: api.payments.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: api
                port:
                  number: 80
```
