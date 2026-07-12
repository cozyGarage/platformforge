# Expose a Deployment with a Service

Deployments alone are not reachable inside a cluster. A **Service** provides stable networking — Boot.dev **Learn Kubernetes — Services**.

A Deployment `api` is already running. Your job is to expose it.

## Tasks

1. Write `/workspace/service.yaml` for a `ClusterIP` Service named `api`
2. Select pods with `app: api`, expose port `80`
3. Run `kubectl apply -f /workspace/service.yaml`

## Starter manifest

```yaml
apiVersion: v1
kind: Service
metadata:
  name: api
spec:
  selector:
    app: api
  ports:
    - port: 80
      targetPort: 80
```
