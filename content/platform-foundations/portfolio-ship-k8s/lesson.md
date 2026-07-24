# Portfolio — ship to Kubernetes

The demo story is incomplete until the service runs behind a Service and Ingress. This lab uses `nginx:1.27-alpine` as a stand-in image so the release path works without a private registry.

## Tasks

1. `/workspace/k8s/deployment.yaml` — Deployment `payments`, `app=payments`, 2 replicas
2. `/workspace/k8s/service.yaml` — ClusterIP Service `payments` port 80
3. `/workspace/k8s/ingress.yaml` — host `payments.local` → Service `payments:80`
4. Apply all manifests and wait for Ready pods
