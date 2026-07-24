# Persist data with PVC and volume mounts

Pods are ephemeral. CKA storage questions usually ask you to claim capacity and mount it into a workload.

## Tasks

1. Create PVC `api-data` for `1Gi` `ReadWriteOnce` → `/workspace/pvc.yaml`
2. Mount it into Deployment `api` at `/data` → `/workspace/deployment.yaml`
3. Apply both and wait for the pod to become Ready

## Example PVC

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: api-data
spec:
  accessModes: [ReadWriteOnce]
  resources:
    requests:
      storage: 1Gi
```
