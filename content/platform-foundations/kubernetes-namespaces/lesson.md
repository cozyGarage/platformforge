# Isolate workloads with namespaces

Namespaces partition a cluster so teams and environments do not collide. Boot.dev **Learn Kubernetes** introduces namespaces as the unit of isolation for Deployments, Services, and ConfigMaps.

## Why namespaces matter

Platform engineers put payment services, staging apps, and shared tooling in separate namespaces. `kubectl get pods` without `-n` only shows the current namespace — usually `default` — which hides mistakes if you deploy to the wrong place.

## Starting state

- An `api` Deployment already runs in `default` (leave it alone)
- You will create a new namespace for a ledger service

## Tasks

1. Create namespace `payments` and save the manifest to `/workspace/namespace.yaml`
2. Create Deployment `ledger` in `payments` using `nginx:1.27-alpine` with label `app=ledger`
3. Save the Deployment manifest to `/workspace/deployment.yaml` (must include `namespace: payments`)
4. Apply and wait until the ledger pod is ready in `payments`

## Example manifests

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: payments
```

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ledger
  namespace: payments
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ledger
  template:
    metadata:
      labels:
        app: ledger
    spec:
      containers:
        - name: ledger
          image: nginx:1.27-alpine
```
