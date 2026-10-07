# Portfolio — Build, Push, and Ship to Kubernetes

The demo story is incomplete until **your** image is in a registry and running behind a Service and Ingress. PlatformForge creates a k3d registry and performs the host-side `docker build` / `docker push` (learners never get a Docker socket).

## Why it matters

Shipping `nginx` as a stand-in skips the hard part. Digests and registry refs are what FinTech platforms actually promote.

## Ticket focus

1. Finish `/workspace/app/Dockerfile` (pinned, non-root)
2. Validate → host builds/pushes → `/workspace/IMAGE`
3. Deploy Deployment/Service/Ingress using that image

## Tip codes

- `REG_DOCKERFILE` — pin golang:1.22-alpine; never latest
- `REG_IMAGE` — deploy `$(cat /workspace/IMAGE)`, not nginx
