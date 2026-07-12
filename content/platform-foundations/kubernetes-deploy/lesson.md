# Deploy a reliable workload to Kubernetes

You inherited a minimal `Deployment` manifest for an `api` service. It runs, but it is missing the safeguards platform teams expect before production: health probes and resource requests/limits.

A live **k3d** cluster is provisioned when you start this lab. Your kubeconfig is already mounted at `/workspace/.kube/config`.

## Your mission

1. Edit `/workspace/deployment.yaml`
2. Add **readiness** and **liveness** HTTP probes (path `/`, port `80`)
3. Add **CPU and memory** requests and limits
4. Apply the manifest: `kubectl apply -f /workspace/deployment.yaml`
5. Confirm pods become ready: `kubectl get pods -l app=api`

## Why this matters

Platform engineers ship guardrails, not just YAML. Probes and resource limits prevent cascading failures and noisy-neighbor issues in shared clusters.

## Solution pattern

```yaml
readinessProbe:
  httpGet:
    path: /
    port: 80
livenessProbe:
  httpGet:
    path: /
    port: 80
resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 100m
    memory: 128Mi
```
