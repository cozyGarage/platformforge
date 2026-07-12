# Scale a deployment horizontally

Kubernetes scales stateless APIs by increasing **replicas**. Boot.dev's **Learn Kubernetes** module covers Deployments; this lab focuses on the operational loop — scale, verify, persist the manifest.

## Starting state

- Deployment `api` runs **1** nginx pod in your k3d cluster

## Tasks

1. Scale `api` to **3** replicas
2. Export the Deployment manifest to `/workspace/deployment.yaml`
3. Wait until every pod reports `Ready`

## Why it matters

Platform teams scale before traffic spikes. You should always confirm pod readiness, not just replica count in the spec.
