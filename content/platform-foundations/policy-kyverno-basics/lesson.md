# Plan Kyverno Baseline Policies

Kyverno admits (or denies) Kubernetes objects with policies that read like YAML you already know.

## Why it matters

Without admission baselines, unlabeled workloads and `:latest` tags slip into prod and break audits, rollbacks, and GitOps digests.

## Ticket focus

1. `require-app-label` ClusterPolicy
2. `block-latest-tag` ClusterPolicy
3. Short POLICY.md summary

## Tip codes

- `KYN_LABEL` — validate `metadata.labels.app`
- `KYN_ENFORCE` — prefer Enforce for prod baselines
- `NO_LATEST_TAG` — deny container images tagged `:latest`
