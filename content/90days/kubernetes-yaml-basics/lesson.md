# Author a first Kubernetes manifest

> **Attribution:** Scenario adapted from [90DaysOfDevOps](https://github.com/MichaelCade/90DaysOfDevOps) Kubernetes fundamentals (Michael Cade, CC BY-NC-SA 4.0). Rewritten as an interactive PlatformForge exercise.

Before live clusters, 90Days teaches desired-state YAML. This lab builds a Deployment you could later apply with `kubectl`.

## Tasks

1. Write `/workspace/k8s/web.yaml`:
   - Deployment `web`
   - label `app=web`
   - `replicas: 1`
   - image `nginx:1.27-alpine`
   - `containerPort: 80`
2. Write `/workspace/k8s/README.md` mentioning `kubectl apply`

## Source

- Original curriculum: https://github.com/MichaelCade/90DaysOfDevOps
- Modifications: single Deployment authoring exercise with deterministic file checks
