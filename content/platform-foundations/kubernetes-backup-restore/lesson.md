# Backup and restore cluster objects

CKA backup topics include etcd snapshots and application-level restores. This lab drills the application pattern: export YAML, survive a delete, restore, and document the drill.

## Starting state

Namespace `payments` has ConfigMap `api-config` (`LOG_LEVEL=info`, `REGION=eu-west-1`) and Deployment `api`.

## Tasks

1. Backup: `kubectl -n payments get cm api-config -o yaml > /workspace/backup/api-config.yaml`
2. Simulate loss: delete the ConfigMap
3. Restore with `kubectl apply -f`
4. Write `/workspace/backup/RESTORE.md` describing the restore and that backups must be tested
