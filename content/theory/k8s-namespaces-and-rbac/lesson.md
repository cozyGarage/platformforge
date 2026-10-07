Kubernetes **namespaces** provide organizational and resource-quota isolation between teams, environments, or tenants sharing a cluster. **RBAC** (Role-Based Access Control) enforces least-privilege access: who can do what to which resources in which namespaces. Together they form the foundation of multi-tenant Kubernetes security.

**Learning objectives**
- Model the RBAC engine: PolicyRule, Role, ClusterRole, RoleBinding
- Implement the RBAC access check algorithm (allow-only, no deny rules)
- Model ResourceQuota enforcement to prevent noisy-neighbor problems
- Simulate a multi-tenant cluster with team isolation

## Core Concepts

| Concept | Definition | Scope |
|---|---|---|
| Namespace | Virtual cluster partition — isolates names + quotas | Cluster |
| ResourceQuota | Caps total resource usage within a namespace | Namespace |
| LimitRange | Defaults + min/max for individual Pods/Containers | Namespace |
| Role | Set of allowed API operations | Namespace-scoped |
| ClusterRole | Set of allowed API operations | Cluster-wide |
| RoleBinding | Binds a Role to subjects | Namespace-scoped |
| ClusterRoleBinding | Binds a ClusterRole to subjects | Cluster-wide |
| ServiceAccount | Identity for Pods to call the Kubernetes API | Namespace |
| Subjects | Users, Groups, or ServiceAccounts in bindings | — |
| Least privilege | Grant only minimum required permissions | Security principle |

```python
from dataclasses import dataclass, field
from typing import List, Dict, Set, Optional, Tuple
from enum import Enum

# ---------------------------------------------------------------------------
# RBAC Data Model
# ---------------------------------------------------------------------------

class Verb(Enum):
    GET    = 'get'
    LIST   = 'list'
    WATCH  = 'watch'
    CREATE = 'create'
    UPDATE = 'update'
    PATCH  = 'patch'
    DELETE = 'delete'

READ_VERBS  = {Verb.GET, Verb.LIST, Verb.WATCH}
WRITE_VERBS = {Verb.CREATE, Verb.UPDATE, Verb.PATCH, Verb.DELETE}
ALL_VERBS   = set(Verb)


@dataclass
class PolicyRule:
    api_groups: List[str]   # '' = core API; 'apps', 'batch', etc.
    resources:  List[str]   # 'pods', 'deployments', '*' for all
    verbs:      Set[Verb]

    def allows(self, api_group: str, resource: str, verb: Verb) -> bool:
        group_match    = '*' in self.api_groups or api_group in self.api_groups
        resource_match = '*' in self.resources  or resource  in self.resources
        return group_match and resource_match and verb in self.verbs


@dataclass
class Role:
    name:      str
    namespace: str
    rules:     List[PolicyRule]


@dataclass
class ClusterRole:
    name:  str
    rules: List[PolicyRule]


@dataclass
class RoleBinding:
    name:      str
    namespace: str
    role_ref:  str           # name of Role or ClusterRole
    is_cluster_role: bool    # True → role_ref is a ClusterRole
    subjects:  List[Dict]    # [{'kind': 'User', 'name': 'alice'}]


@dataclass
class ClusterRoleBinding:
    name:        str
    role_ref:    str
    subjects:    List[Dict]


@dataclass
class RBACEngine:
    """
    Implements Kubernetes RBAC: allow-only, no explicit deny.
    Access = ALLOWED iff at least one rule in a bound role grants it.
    """
    roles:                Dict[str, Role]              = field(default_factory=dict)
    cluster_roles:        Dict[str, ClusterRole]       = field(default_factory=dict)
    role_bindings:        List[RoleBinding]            = field(default_factory=list)
    cluster_role_bindings: List[ClusterRoleBinding]    = field(default_factory=list)

    def _subject_match(self, subjects: List[Dict], username: str) -> bool:
        return any(s.get('name') == username for s in subjects)

    def can(self, username: str, namespace: str, api_group: str, resource: str, verb: Verb) -> bool:
        # 1. Check namespace-scoped RoleBindings
        for rb in self.role_bindings:
            if rb.namespace != namespace:
                continue
            if not self._subject_match(rb.subjects, username):
                continue
            if rb.is_cluster_role:
                role_rules = self.cluster_roles.get(rb.role_ref, ClusterRole('', [])).rules
            else:
                key = f'{namespace}/{rb.role_ref}'
                role_rules = self.roles.get(key, Role('', '', [])).rules
            if any(r.allows(api_group, resource, verb) for r in role_rules):
                return True

        # 2. Check cluster-wide ClusterRoleBindings
        for crb in self.cluster_role_bindings:
            if not self._subject_match(crb.subjects, username):
                continue
            cr = self.cluster_roles.get(crb.role_ref, ClusterRole('', []))
            if any(r.allows(api_group, resource, verb) for r in cr.rules):
                return True
        return False

    def audit_user(self, username: str, namespace: str) -> None:
        checks: List[Tuple[str, str, Verb]] = [
            ('',     'pods',          Verb.GET),
            ('',     'pods',          Verb.CREATE),
            ('',     'pods',          Verb.DELETE),
            ('apps', 'deployments',   Verb.GET),
            ('apps', 'deployments',   Verb.UPDATE),
            ('',     'secrets',       Verb.GET),
            ('',     'configmaps',    Verb.GET),
            ('',     'namespaces',    Verb.CREATE),
            ('rbac.authorization.k8s.io', 'roles', Verb.CREATE),
        ]
        print(f'RBAC audit: {username} in namespace={namespace}')
        for api_group, resource, verb in checks:
            allowed = self.can(username, namespace, api_group, resource, verb)
            marker  = 'ALLOW' if allowed else 'DENY '
            print(f'  [{marker}] {verb.value:8} {resource:20} (apiGroup={api_group or "core"})')


# ---------------------------------------------------------------------------
# Build a realistic multi-team RBAC configuration
# ---------------------------------------------------------------------------
rbac = RBACEngine()

# Developer role: read most things, manage pods, no secrets access
rbac.roles['production/developer'] = Role('developer', 'production', [
    PolicyRule(['', 'apps'], ['pods', 'deployments', 'replicasets', 'services', 'configmaps'],
               READ_VERBS),
    PolicyRule([''], ['pods'], {Verb.CREATE, Verb.DELETE}),
])

# Ops cluster role: full cluster access
rbac.cluster_roles['ops-admin'] = ClusterRole('ops-admin', [
    PolicyRule(['*'], ['*'], ALL_VERBS),
])

# Read-only viewer: cluster-wide
rbac.cluster_roles['viewer'] = ClusterRole('viewer', [
    PolicyRule(['*'], ['*'], READ_VERBS),
])

# Bindings
rbac.role_bindings += [
    RoleBinding('alice-dev',  'production', 'developer', False, [{'kind': 'User', 'name': 'alice'}]),
    RoleBinding('carol-dev',  'staging',    'developer', False, [{'kind': 'User', 'name': 'carol'}]),
    # Bob uses ClusterRole bound to a namespace via RoleBinding
    RoleBinding('bob-viewer', 'production', 'viewer',    True,  [{'kind': 'User', 'name': 'bob'}]),
]
rbac.cluster_role_bindings += [
    ClusterRoleBinding('sre-admin', 'ops-admin', [{'kind': 'User', 'name': 'sre'}]),
]

# Audit
rbac.audit_user('alice', 'production')   # developer in production
print()
rbac.audit_user('bob', 'production')     # viewer in production
print()
rbac.audit_user('sre', 'production')     # cluster admin
```

## ResourceQuota and LimitRange Enforcement

```python
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple


@dataclass
class ResourceQuota:
    namespace: str
    max_pods: int = 50
    max_cpu_requests: float = 10.0   # CPU cores
    max_memory_requests_gi: float = 20.0  # GiB
    max_services: int = 20


@dataclass
class NamespaceUsage:
    namespace: str
    current_pods: int = 0
    current_cpu_requests: float = 0.0   # CPU cores
    current_memory_gi: float = 0.0
    current_services: int = 0

    def can_schedule_pod(
        self,
        quota: ResourceQuota,
        cpu_request_cores: float,
        memory_request_gi: float,
    ) -> Tuple[bool, str]:
        if self.current_pods + 1 > quota.max_pods:
            return False, f'Pod quota exceeded ({self.current_pods}/{quota.max_pods})'
        if self.current_cpu_requests + cpu_request_cores > quota.max_cpu_requests:
            return False, (f'CPU quota exceeded: would be '
                           f'{self.current_cpu_requests + cpu_request_cores:.1f}'
                           f'/{quota.max_cpu_requests} cores')
        if self.current_memory_gi + memory_request_gi > quota.max_memory_requests_gi:
            return False, (f'Memory quota exceeded: would be '
                           f'{self.current_memory_gi + memory_request_gi:.1f}'
                           f'/{quota.max_memory_requests_gi} GiB')
        return True, 'Admitted'

    def admit_pod(self, cpu_cores: float, memory_gi: float) -> None:
        self.current_pods += 1
        self.current_cpu_requests += cpu_cores
        self.current_memory_gi += memory_gi

    def utilization(self, quota: ResourceQuota) -> Dict[str, str]:
        return {
            'pods':   f'{self.current_pods}/{quota.max_pods}',
            'cpu':    f'{self.current_cpu_requests:.1f}/{quota.max_cpu_requests} cores '
                      f'({self.current_cpu_requests/quota.max_cpu_requests*100:.0f}%)',
            'memory': f'{self.current_memory_gi:.1f}/{quota.max_memory_requests_gi} GiB '
                      f'({self.current_memory_gi/quota.max_memory_requests_gi*100:.0f}%)',
        }


# Simulate two teams sharing a cluster with per-namespace quotas
quotas = {
    'team-ml':      ResourceQuota('team-ml',      max_pods=30, max_cpu_requests=16.0, max_memory_requests_gi=32.0),
    'team-backend': ResourceQuota('team-backend',  max_pods=20, max_cpu_requests=8.0,  max_memory_requests_gi=16.0),
}
usage = {
    'team-ml':      NamespaceUsage('team-ml'),
    'team-backend': NamespaceUsage('team-backend'),
}

# Simulate pod scheduling requests
pod_requests = [
    ('team-ml',      0.5,  1.0,  'ml-worker-01'),
    ('team-ml',      0.5,  1.0,  'ml-worker-02'),
    ('team-ml',      4.0,  8.0,  'ml-gpu-job-01'),      # large GPU job
    ('team-backend', 0.2,  0.5,  'api-pod-01'),
    ('team-backend', 0.2,  0.5,  'api-pod-02'),
    ('team-ml',      8.0, 16.0,  'ml-gpu-job-02'),      # will exceed quota
    ('team-backend', 8.0, 16.0,  'backend-large'),      # will exceed quota
]

print('ResourceQuota enforcement simulation:')
print(f'{"Namespace":>15}  {"Pod":>18}  {"CPU":>6}  {"Mem":>6}  Result')
print('-' * 75)
for ns, cpu, mem, name in pod_requests:
    allowed, reason = usage[ns].can_schedule_pod(quotas[ns], cpu, mem)
    if allowed:
        usage[ns].admit_pod(cpu, mem)
    status = 'ADMITTED' if allowed else f'REJECTED: {reason}'
    print(f'  {ns:>13}  {name:>18}  {cpu:>5.1f}c  {mem:>5.1f}G  {status}')

print()
print('Final namespace utilization:')
for ns, u in usage.items():
    util = u.utilization(quotas[ns])
    print(f'  {ns}:')
    for k, v in util.items():
        print(f'    {k:8}: {v}')
```

## Quick Quiz

**Q1**: What is the difference between a Role and a ClusterRole, and when would you use a ClusterRoleBinding vs a RoleBinding?
> **A**: Role is namespace-scoped (only grants permissions within one namespace). ClusterRole is cluster-wide. A RoleBinding can bind either a Role or ClusterRole to subjects, but the access is scoped to the RoleBinding's namespace. A ClusterRoleBinding binds a ClusterRole cluster-wide — use this for cluster-admin or metrics-server that needs to read resources in all namespaces.

**Q2**: Why are namespaces not sufficient as a security boundary?
> **A**: Namespace isolation is purely organizational — Pods can communicate across namespaces by default. Without NetworkPolicies, any pod can reach any other pod's IP in the cluster. RBAC only restricts Kubernetes API calls, not network traffic.

**Q3**: What happens if you grant a ServiceAccount `verbs: [*]` on `resources: [roles, rolebindings]`?
> **A**: The pod can create new Roles with any permissions and bind them to itself, effectively escalating to cluster-admin. This is a privilege escalation vulnerability. Kubernetes has safeguards (`bind` and `escalate` verbs) but only if `--enable-bootstrap-token-auth` RBAC settings are correct.

## Key Takeaways

- RBAC is allow-only: no matching rule = denied; no explicit deny needed
- Role = namespace scope; ClusterRole = cluster scope; RoleBinding scopes either to a namespace
- ServiceAccounts provide pod identity — grant them only the API calls their workload needs
- ResourceQuotas prevent noisy-neighbor resource starvation between teams
- Namespaces are organizational, not security boundaries — add NetworkPolicy for network isolation

## Going Deeper

1. **Exercise**: Add `impersonation` support to `RBACEngine` — `kubectl --as=alice` allows an admin to act as alice. Model the `impersonate` verb on `users` resource and enforce it.
2. **Exercise**: Implement `NetworkPolicy` simulation — model allow/deny rules for ingress/egress and show which pods can communicate with which.
3. **Exercise**: Audit for privilege escalation: scan all RoleBindings for any that grant `create/update` on `roles` or `rolebindings` to non-admin subjects.
4. **Project**: Build a `kubectl auth can-i` simulator that takes `(user, namespace, verb, resource)` and returns allowed/denied with the matching rule path.
5. **Research**: What is OPA (Open Policy Agent) Gatekeeper and how does it extend beyond RBAC? Compare to Kyverno — what policy use cases does admission control cover that RBAC cannot?

## Hands-on Lab

A runnable companion lab lives in `lab_04_namespaces_rbac/`. It provisions a local k3d cluster, three namespaces (`dev`, `staging`, `prod`), three `ServiceAccount`s, a `ClusterRole` (`pod-reader`), namespace-scoped `Role`s (`editor`, `admin`), and the matching `RoleBinding`s.

Walk-through:

```bash
cd lab_04_namespaces_rbac
make up              # k3d cluster + apply manifests
make test-rbac       # auth can-i matrix: SA x verb x resource x namespace
make kubeconfig-dev  # mint a token kubeconfig for dev-readonly
make kubeconfig-prod # mint a token kubeconfig for prod-admin
make down
```

What you will demonstrate:

- `dev-readonly` cannot delete pods (`kubectl auth can-i delete pods` -> no).
- `prod-admin` can delete pods in `prod` but not in `dev` or `staging`.
- A `ClusterRole` referenced from a `RoleBinding` is scoped to that namespace only (vs. a `ClusterRoleBinding`, which would grant it cluster-wide).
- `scripts/exec-as.sh` uses the `TokenRequest` API to mint a short-lived bearer token and writes a self-contained kubeconfig so you can run commands as that ServiceAccount.

This makes the RBAC engine modeled above observable on a real API server.
