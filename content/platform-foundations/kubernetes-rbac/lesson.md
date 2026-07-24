# Grant least-privilege RBAC

CKA RBAC questions reward narrow Roles: only the verbs and resources a workload needs, bound in one namespace.

## Tasks

1. Create ServiceAccount `deployer` in `payments` → `/workspace/sa.yaml`
2. Create Role `deployer` allowing Deployment get/list/watch/create/update/patch → `/workspace/role.yaml`
3. Bind them with RoleBinding → `/workspace/rolebinding.yaml`
4. Apply all three and verify with `kubectl auth can-i`

## Verify

```bash
kubectl auth can-i create deployments --as=system:serviceaccount:payments:deployer -n payments
kubectl auth can-i get secrets --as=system:serviceaccount:payments:deployer -n payments
```

Expect `yes` for Deployments and `no` for Secrets.
