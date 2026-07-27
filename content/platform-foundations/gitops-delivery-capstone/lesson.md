# Capstone — Ship a GitOps Delivery Pack

`payments-web` is stuck in a broken draft: floating `latest`, no prod values, a plaintext Secret in Git, and no rotation story. Build a shippable delivery pack.

## Release checklist

1. Harden the Helm chart and pin prod digests
2. Declare GitOps desired state + reconcile policy
3. Replace plaintext Secret with ExternalSecret
4. Write Detect → Rotate → Refresh → Verify plus a three-point checklist
5. Leave a short ship note and commit clean

This capstone combines Helm packaging, GitOps sync planning, External Secrets, and rotation runbooks.
