# Apply Kyverno Baseline Policies

Kyverno admits (or denies) Kubernetes objects with policies that read like YAML you already know — and this lab applies them on a live k3d cluster.

## Why it matters

Planning YAML is not enough. Admission only protects the cluster after `kubectl apply` and the webhook is Ready.

## Ticket focus

1. ClusterPolicy `require-app-label` with Enforce
2. ClusterPolicy `block-latest-tag`
3. Apply both; confirm with `kubectl get clusterpolicy`

## Tip codes

- `KYN_LABEL` — validate metadata.labels.app exists
- `KYN_ENFORCE` — validationFailureAction: Enforce
- `NO_LATEST_TAG` — deny images tagged :latest
