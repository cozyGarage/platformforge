# Practice etcd Snapshot Runbooks

CKA disaster recovery expects `etcdctl snapshot save` / `restore`. This lab gives you a **real snapshot** taken from the k3d control-plane etcd (host-side, no privileged lab container) plus the companion manifest habit.

## Why it matters

Runbooks without a verifiable snapshot artifact are theater. Inspect status, checksum the bytes, and keep app-consistent manifests beside etcd.

## Ticket focus

1. `etcdctl snapshot status` on the mounted live `snapshot.db`
2. sha256 of the live snapshot + ConfigMap companion export
3. Snapshot / Restore / Validation runbook

## Tip codes

- `ETCD_STATUS` — snapshot status on the mounted control-plane artifact
