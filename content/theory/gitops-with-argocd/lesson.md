**GitOps** treats Git as the single source of truth for cluster state. **ArgoCD** is the Kubernetes-native GitOps controller that continuously reconciles the actual cluster state with the desired state defined in Git — detecting drift, enabling auditable rollbacks via `git revert`, and automating deployments through pull requests.

**Learning objectives**
- Model the ArgoCD Application CRD and its sync/health status machine
- Simulate the reconciliation loop: diff desired vs actual, apply deltas
- Implement drift detection and self-heal mechanics
- Model multi-environment promotion with ApplicationSets

## Core Concepts

| Concept | Definition | Key Detail |
|---|---|---|
| ArgoCD Application | CRD linking Git repo/path to cluster/namespace | The unit of GitOps management |
| Sync status | Synced / OutOfSync / Unknown | Cluster matches Git or not |
| Health status | Healthy / Progressing / Degraded / Missing | Actual workload health |
| OutOfSync | Git differs from cluster state | Detected via hash comparison |
| Sync | Apply Git state to cluster | Manual or automatic |
| Auto-sync | ArgoCD syncs on Git change | Continuous delivery |
| Prune | Delete orphaned cluster resources | Resources absent from Git |
| Self-heal | Revert manual cluster changes | Enforces Git as only truth |
| ApplicationSet | Template generating multiple Applications | Multi-env, multi-cluster |
| Sync waves | Ordered sync of resource groups | Deploy CRDs before CRs |
| App-of-apps | ArgoCD app managing other ArgoCD apps | Bootstrapping pattern |

```python
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Set
from enum import Enum
import hashlib
import json

# ---------------------------------------------------------------------------
# ArgoCD core types
# ---------------------------------------------------------------------------

class SyncStatus(Enum):
    SYNCED       = 'Synced'
    OUT_OF_SYNC  = 'OutOfSync'
    UNKNOWN      = 'Unknown'


class HealthStatus(Enum):
    HEALTHY      = 'Healthy'
    PROGRESSING  = 'Progressing'
    DEGRADED     = 'Degraded'
    MISSING      = 'Missing'
    SUSPENDED    = 'Suspended'


@dataclass(frozen=True)
class K8sResource:
    """
    Immutable representation of a Kubernetes resource.
    Identity: (kind, namespace, name). Content: spec hash.
    """
    kind:      str
    name:      str
    namespace: str
    spec:      str  # JSON-serialized spec

    @property
    def key(self) -> str:
        return f'{self.kind}/{self.namespace}/{self.name}'

    @property
    def content_hash(self) -> str:
        return hashlib.sha256(self.spec.encode()).hexdigest()[:8]

    def __repr__(self) -> str:
        return f'{self.kind}/{self.name}@{self.content_hash}'


def make_resource(kind: str, name: str, ns: str, spec: dict) -> K8sResource:
    return K8sResource(kind, name, ns, json.dumps(spec, sort_keys=True))


@dataclass
class ResourceDiff:
    creates: List[K8sResource] = field(default_factory=list)
    updates: List[K8sResource] = field(default_factory=list)
    deletes: List[K8sResource] = field(default_factory=list)

    @property
    def is_empty(self) -> bool:
        return not (self.creates or self.updates or self.deletes)

    def show(self) -> None:
        if self.is_empty:
            print('  (no diff — already synced)')
            return
        for r in self.creates:
            print(f'  + CREATE {r}')
        for r in self.updates:
            print(f'  ~ UPDATE {r}')
        for r in self.deletes:
            print(f'  - DELETE {r}')


@dataclass
class ArgoCDApp:
    name:             str
    repo_url:         str
    target_revision:  str   # branch or commit SHA
    path:             str   # path within repo
    dest_namespace:   str
    auto_sync:        bool = False
    prune:            bool = False
    self_heal:        bool = False
    _desired: Dict[str, K8sResource] = field(default_factory=dict)  # from Git
    _actual:  Dict[str, K8sResource] = field(default_factory=dict)  # from cluster
    _sync_history: List[Dict] = field(default_factory=list)

    # -------------------------
    # Status properties
    # -------------------------
    @property
    def sync_status(self) -> SyncStatus:
        if not self._desired:
            return SyncStatus.UNKNOWN
        return SyncStatus.SYNCED if self._diff().is_empty else SyncStatus.OUT_OF_SYNC

    @property
    def health_status(self) -> HealthStatus:
        # Simple model: healthy if synced and all expected resources exist
        if not self._actual:
            return HealthStatus.MISSING
        if self.sync_status == SyncStatus.OUT_OF_SYNC:
            return HealthStatus.PROGRESSING
        return HealthStatus.HEALTHY

    # -------------------------
    # Core operations
    # -------------------------
    def load_from_git(self, resources: List[K8sResource], commit_sha: str = '') -> None:
        """Simulate ArgoCD polling Git and loading desired state."""
        self._desired = {r.key: r for r in resources}
        if commit_sha:
            self.target_revision = commit_sha

    def _diff(self) -> ResourceDiff:
        diff = ResourceDiff()
        for key, desired in self._desired.items():
            actual = self._actual.get(key)
            if actual is None:
                diff.creates.append(desired)
            elif actual.content_hash != desired.content_hash:
                diff.updates.append(desired)
        if self.prune:
            for key, actual in self._actual.items():
                if key not in self._desired:
                    diff.deletes.append(actual)
        return diff

    def sync(self, force: bool = False) -> ResourceDiff:
        diff = self._diff()
        if diff.is_empty and not force:
            return diff
        # Apply diff
        new_actual = dict(self._actual)
        for r in diff.creates + diff.updates:
            new_actual[r.key] = r
        for r in diff.deletes:
            new_actual.pop(r.key, None)
        self._actual = new_actual
        self._sync_history.append({
            'revision': self.target_revision,
            'creates': len(diff.creates),
            'updates': len(diff.updates),
            'deletes': len(diff.deletes),
        })
        return diff

    def simulate_drift(self, key: str, patched_spec: dict) -> None:
        """Someone ran kubectl edit — mutate actual state directly."""
        if key in self._actual:
            old = self._actual[key]
            drifted = K8sResource(old.kind, old.name, old.namespace,
                                   json.dumps(patched_spec, sort_keys=True))
            self._actual[key] = drifted

    def status_line(self) -> str:
        return (f'  {self.name}: sync={self.sync_status.value:12} '
                f'health={self.health_status.value:12} '
                f'resources={len(self._actual)}')


# ---------------------------------------------------------------------------
# Simulate a full ArgoCD application lifecycle
# ---------------------------------------------------------------------------
app = ArgoCDApp(
    name='ml-inference-api',
    repo_url='https://github.com/org/k8s-manifests',
    target_revision='main',
    path='apps/ml-inference-api',
    dest_namespace='production',
    auto_sync=True, prune=True, self_heal=True,
)

# --- Step 1: Initial deploy (Git has v1.0 manifests) ---
v1_resources = [
    make_resource('Deployment',  'ml-api',         'production', {'image': 'ml-api:1.0', 'replicas': 3}),
    make_resource('Service',     'ml-api-svc',     'production', {'port': 80, 'type': 'ClusterIP'}),
    make_resource('ConfigMap',   'ml-api-config',  'production', {'MODEL': 'resnet50', 'BATCH_SIZE': '32'}),
]
app.load_from_git(v1_resources, commit_sha='abc1234')
print('Step 1: Initial load from Git')
print(f'  Sync status before sync: {app.sync_status.value}')
diff = app.sync()
print('  Diff applied:')
diff.show()
print(app.status_line())

print()
# --- Step 2: Developer pushes new image + HPA ---
v2_resources = [
    make_resource('Deployment',  'ml-api',         'production', {'image': 'ml-api:1.1', 'replicas': 3}),
    make_resource('Service',     'ml-api-svc',     'production', {'port': 80, 'type': 'ClusterIP'}),
    make_resource('ConfigMap',   'ml-api-config',  'production', {'MODEL': 'resnet50', 'BATCH_SIZE': '64'}),
    make_resource('HPA',         'ml-api-hpa',     'production', {'min': 3, 'max': 15, 'target_cpu': 70}),
]
app.load_from_git(v2_resources, commit_sha='def5678')
print('Step 2: git push — image v1.1 + HPA + config change')
print(f'  Sync status: {app.sync_status.value}')
diff = app.sync()
print('  Diff applied:')
diff.show()
print(app.status_line())

print()
# --- Step 3: Operator manually scales deployment (drift) ---
print('Step 3: Operator ran: kubectl scale deploy/ml-api --replicas=10')
app.simulate_drift('Deployment/production/ml-api', {'image': 'ml-api:1.1', 'replicas': 10})
print(f'  Drift detected: sync={app.sync_status.value}')

# Self-heal: ArgoCD detects drift and re-syncs
if app.self_heal and app.sync_status == SyncStatus.OUT_OF_SYNC:
    print('  Self-heal triggered: reverting to Git state')
    diff = app.sync()
    diff.show()
    print(app.status_line())
```

## ApplicationSet — Multi-Environment Promotion

```python
from dataclasses import dataclass, field
from typing import Dict, List, Any

@dataclass
class AppSetGenerator:
    """List generator — simplest ApplicationSet generator type."""
    elements: List[Dict[str, Any]]


@dataclass
class AppSetTemplate:
    """Template for generating ArgoCD Applications."""
    name_pattern: str          # e.g. 'ml-api-{{env}}'
    repo_url: str
    path_pattern: str          # e.g. 'envs/{{env}}'
    dest_namespace_pattern: str  # e.g. '{{env}}'
    auto_sync: bool = False
    prune: bool = True


@dataclass
class ApplicationSet:
    name: str
    generator: AppSetGenerator
    template: AppSetTemplate

    def generate_applications(self) -> List[ArgoCDApp]:
        apps = []
        for elem in self.generator.elements:
            def render(pattern: str, params: dict) -> str:
                result = pattern
                for k, v in params.items():
                    result = result.replace(f'{{{{{k}}}}}', str(v))
                return result

            apps.append(ArgoCDApp(
                name=render(self.template.name_pattern, elem),
                repo_url=self.template.repo_url,
                target_revision=elem.get('revision', 'main'),
                path=render(self.template.path_pattern, elem),
                dest_namespace=render(self.template.dest_namespace_pattern, elem),
                auto_sync=elem.get('auto_sync', self.template.auto_sync),
                prune=self.template.prune,
            ))
        return apps


# Define ApplicationSet for ml-api across environments
appset = ApplicationSet(
    name='ml-api-environments',
    generator=AppSetGenerator([
        {'env': 'dev',        'revision': 'main',    'auto_sync': True},
        {'env': 'staging',    'revision': 'main',    'auto_sync': True},
        {'env': 'production', 'revision': 'v1.5.0',  'auto_sync': False},  # manual in prod
    ]),
    template=AppSetTemplate(
        name_pattern='ml-api-{{env}}',
        repo_url='https://github.com/org/k8s-manifests',
        path_pattern='envs/{{env}}/ml-api',
        dest_namespace_pattern='{{env}}',
        auto_sync=False,
        prune=True,
    ),
)

generated_apps = appset.generate_applications()
print(f'ApplicationSet "{appset.name}" generated {len(generated_apps)} Applications:')
print(f'{"Name":>25}  {"Namespace":>12}  {"Revision":>10}  {"AutoSync"}')
print('-' * 65)
for app_gen in generated_apps:
    print(f'  {app_gen.name:>23}  {app_gen.dest_namespace:>12}  '
          f'{app_gen.target_revision:>10}  {app_gen.auto_sync}')

# Simulate promotion: dev auto-syncs on every commit, staging after PR merge,
# production requires manual ArgoCD sync
print()
print('Promotion model:')
print('  dev        → auto-sync on every commit to main')
print('  staging    → auto-sync on merge to main (same branch)')
print('  production → manual sync after staging validation, pinned to release tag')

# Simulate sync events
resources_v1 = [
    make_resource('Deployment', 'ml-api', env_app.dest_namespace,
                  {'image': f'ml-api:1.5.0', 'replicas': 2})
    for env_app in generated_apps
]

print()
print('Sync all environments:')
for i, env_app in enumerate(generated_apps):
    env_app.load_from_git([resources_v1[i]], commit_sha='sha_v1.5.0')
    diff = env_app.sync()
    print(env_app.status_line())
```

## Quick Quiz

**Q1**: What is the difference between `prune: true` and `self_heal: true` in ArgoCD?
> **A**: `prune: true` deletes cluster resources that exist in the cluster but are absent from Git (orphaned resources). `self_heal: true` reverts manual changes to resources that ARE in Git — someone ran `kubectl edit` and ArgoCD reverts the change on next reconcile.

**Q2**: How does ArgoCD handle the case where a sync partially fails (3/5 resources applied)?
> **A**: ArgoCD reports the application as `OutOfSync` with health `Degraded`. The 3 applied resources stay in the cluster; the 2 failed resources remain as their old versions. ArgoCD will retry the sync based on retry policy settings (`syncPolicy.retry`). The partial state is visible in the ArgoCD UI per-resource.

**Q3**: How do sync waves help with resource ordering?
> **A**: Resources annotated with `argocd.argoproj.io/sync-wave: "0"` sync before wave 1, which syncs before wave 2. This enables: CRDs (wave 0) → Namespaces (wave 1) → Deployments (wave 2). ArgoCD waits for all resources in wave N to be healthy before proceeding to wave N+1.

## Key Takeaways

- GitOps = Git as the single source of truth; ArgoCD continuously reconciles cluster to match Git
- `OutOfSync` means Git differs from cluster — auto-sync applies Git state; manual sync requires approval
- `prune` removes orphaned resources; `self_heal` reverts manual cluster changes
- ApplicationSets generate multiple Applications from a template — enabling multi-env, multi-cluster deployments
- Rollback = `git revert` → ArgoCD auto-syncs the reverted state — no kubectl required

## Going Deeper

1. **Exercise**: Implement sync waves in `ArgoCDApp.sync()` — annotate resources with wave numbers and enforce ordering by waiting for each wave before proceeding.
2. **Exercise**: Add a `health_check(resource) -> HealthStatus` function that checks deployment rollout status (available replicas vs desired) and propagates health to the Application.
3. **Exercise**: Model the `app-of-apps` pattern — an ArgoCD Application whose Git path contains other ArgoCD Application manifests. Bootstrap a new cluster from a single seed application.
4. **Project**: Build a `GitOpsPromotion` simulator — commit to dev branch triggers dev sync; PR merge to main triggers staging sync; Git tag triggers production sync.
5. **Research**: Compare ArgoCD to Flux CD — architecture differences (operator vs agent model), CRD design, and multi-tenancy security model.

## Hands-on Lab

The simulator above models the ArgoCD reconciliation loop in Python. The companion lab runs the real thing: a local k3d cluster with ArgoCD installed from upstream, two `Application` CRs (`argoproj.io/v1alpha1`) pointing at the canonical `argocd-example-apps` repo, and a scripted drift scenario where you manually scale a managed Deployment and watch ArgoCD revert it — exactly the `self_heal` path simulated in Step 3 above.

Cluster boot to UI in under 5 minutes on an M1 with 8 GB RAM.

- `make up` — k3d cluster + ArgoCD install
- `make apply-apps` — register `guestbook` and `helm-guestbook` Applications
- `make argocd-ui` — print admin password, port-forward UI to `https://localhost:8080`
- `make drift-demo` — `kubectl scale` a managed Deployment to introduce drift
- Watch the ArgoCD UI flip `OutOfSync` → `Synced` as self-heal reverts the change
- `make down` — destroy the cluster

See [`lab_06_gitops_argocd/README.md`](lab_06_gitops_argocd/README.md) for the full walkthrough, expected output, and how to swap the upstream demo repo for your own fork.
