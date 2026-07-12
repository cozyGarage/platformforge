# Inject configuration with ConfigMaps

Hard-coding config in images is an anti-pattern. ConfigMaps decouple configuration from container images — Boot.dev **Learn Kubernetes — ConfigMaps**.

## Tasks

1. Create `/workspace/configmap.yaml` defining `api-config` with `LOG_LEVEL: debug`
2. Patch the `api` Deployment to inject `LOG_LEVEL` from the ConfigMap (save as `/workspace/deployment-patch.yaml` or edit in place)
3. Apply both manifests and verify `kubectl exec deploy/api -- printenv LOG_LEVEL` prints `debug`

## ConfigMap example

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-config
data:
  LOG_LEVEL: debug
```
