# Practice etcd snapshot runbooks

CKA disaster recovery expects you to know `etcdctl snapshot save` / `restore`. This lab builds the runbook habit and a verifiable companion manifest snapshot (full etcd access varies by cluster distro).

## Tasks

1. Export `payments/cluster-meta` ConfigMap to `/workspace/backup/etcd/snapshot-manifests.yaml`
2. Write its sha256 to `/workspace/backup/etcd/SNAPSHOT.sha256`
3. Author `/workspace/docs/etcd-runbook.md` with **Snapshot**, **Restore**, and **Validation**
4. Mention `etcdctl snapshot save`, `etcdctl snapshot restore`, and testing restores
