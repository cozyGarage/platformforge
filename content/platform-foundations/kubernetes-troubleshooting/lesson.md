# Troubleshoot a broken release

CKA troubleshooting is pattern recognition: read events, compare desired vs observed, fix the smallest thing that restores Ready pods and working Services.

## Symptoms in this lab

1. **ImagePullBackOff** — Deployment uses `nginx:bogus-tag`
2. **Empty Endpoints** — Service selects `app=web` but pods are `app=api`

## Tasks

1. Inspect with `kubectl get pods`, `kubectl describe pod`, `kubectl get endpoints api`
2. Set the Deployment image to `nginx:1.27-alpine`
3. Set the Service selector to `app=api`
4. Save `/workspace/deployment.yaml` and `/workspace/service.yaml`, apply, wait for Ready pods
