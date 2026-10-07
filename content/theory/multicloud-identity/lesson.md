## 3.1  The Identity Problem in Multi-Cloud

Each cloud has its own identity system. Without a federation strategy, you end up with:

- Long-lived API keys shared across teams
- Duplicate user accounts on every provider
- No single source of truth for "who has access to what"
- Manual revocation workflows when an employee leaves

**The goal:** one identity plane that spans all clouds, with short-lived credentials and machine identities that require no stored secrets.

## 3.2  Identity Systems by Provider

| Concept | AWS | GCP | Azure |
|---------|-----|-----|-------|
| Human identity | IAM Users / SSO | Cloud Identity | Azure AD (Entra ID) |
| Machine identity | IAM Roles (IRSA, EC2 profiles) | Service Accounts | Managed Identities |
| Federation standard | SAML 2.0, OIDC | OIDC, SAML | SAML, OIDC, WS-Fed |
| Cross-account access | IAM Role assume | Workload Identity Federation | Cross-tenant app registrations |
| Policy language | JSON (IAM policies) | JSON (IAM bindings) | JSON (Azure RBAC) |

## 3.3  Workload Identity Federation (No Stored Secrets)

The modern pattern for machine-to-cloud authentication:

```
1. Workload obtains short-lived token from its native identity provider
   (GitHub Actions OIDC token, Kubernetes service account JWT)

2. Workload presents token to target cloud's STS
   (AWS STS AssumeRoleWithWebIdentity, GCP STS exchangeToken)

3. Cloud validates token signature against trusted OIDC issuer

4. Cloud returns short-lived cloud credentials (15 min–1 hr)

5. Workload uses credentials to call cloud APIs
   → No long-lived secret ever stored in CI/CD or source code
```

```python
# IAM policy evaluator simulation
# Models Allow/Deny evaluation logic matching AWS IAM semantics

from dataclasses import dataclass, field
from typing import Dict, List, Optional
from enum import Enum


class Effect(str, Enum):
    ALLOW = 'Allow'
    DENY  = 'Deny'


@dataclass
class PolicyStatement:
    effect: Effect
    actions: List[str]           # e.g. ['s3:GetObject', 's3:*']
    resources: List[str]         # e.g. ['arn:aws:s3:::my-bucket/*']
    conditions: Dict[str, str] = field(default_factory=dict)


@dataclass
class Policy:
    name: str
    statements: List[PolicyStatement]


@dataclass
class AccessRequest:
    principal: str        # who is asking
    action: str           # e.g. 's3:GetObject'
    resource: str         # e.g. 'arn:aws:s3:::my-bucket/file.txt'
    context: Dict[str, str] = field(default_factory=dict)


def _matches_pattern(pattern: str, value: str) -> bool:
    """Simple wildcard matching: '*' matches anything, '?' matches one char."""
    if pattern == '*':
        return True
    if '*' in pattern:
        prefix, _, suffix = pattern.partition('*')
        return value.startswith(prefix) and value.endswith(suffix)
    return pattern == value


def evaluate_policies(request: AccessRequest, policies: List[Policy]) -> str:
    """
    AWS IAM evaluation logic:
    1. Default DENY
    2. Explicit DENY wins over any ALLOW
    3. Explicit ALLOW required to grant access
    """
    has_allow = False

    for policy in policies:
        for stmt in policy.statements:
            # Check if this statement applies
            action_match   = any(_matches_pattern(a, request.action)   for a in stmt.actions)
            resource_match = any(_matches_pattern(r, request.resource) for r in stmt.resources)

            if not (action_match and resource_match):
                continue

            # Check conditions
            conditions_met = all(
                request.context.get(k) == v
                for k, v in stmt.conditions.items()
            )
            if not conditions_met:
                continue

            if stmt.effect == Effect.DENY:
                return f'DENY (explicit deny in policy: {policy.name})'
            elif stmt.effect == Effect.ALLOW:
                has_allow = True

    if has_allow:
        return 'ALLOW'
    return 'DENY (no matching allow statement — implicit deny)'


# ── Example policies ─────────────────────────────────────────────────────────
read_only_s3 = Policy('ReadOnlyS3', [
    PolicyStatement(
        effect=Effect.ALLOW,
        actions=['s3:GetObject', 's3:ListBucket'],
        resources=['arn:aws:s3:::data-lake/*', 'arn:aws:s3:::data-lake'],
    ),
])

deny_pii_bucket = Policy('DenyPIIAccess', [
    PolicyStatement(
        effect=Effect.DENY,
        actions=['s3:*'],
        resources=['arn:aws:s3:::pii-data/*'],
    ),
])

mfa_required_delete = Policy('MFARequiredForDelete', [
    PolicyStatement(
        effect=Effect.ALLOW,
        actions=['s3:DeleteObject'],
        resources=['arn:aws:s3:::data-lake/*'],
        conditions={'aws:MultiFactorAuthPresent': 'true'},
    ),
])

all_policies = [read_only_s3, deny_pii_bucket, mfa_required_delete]

test_requests = [
    AccessRequest('analyst-role', 's3:GetObject',   'arn:aws:s3:::data-lake/report.parquet'),
    AccessRequest('analyst-role', 's3:GetObject',   'arn:aws:s3:::pii-data/customers.csv'),
    AccessRequest('analyst-role', 's3:PutObject',   'arn:aws:s3:::data-lake/output.csv'),
    AccessRequest('analyst-role', 's3:DeleteObject','arn:aws:s3:::data-lake/old.parquet'),
    AccessRequest('analyst-role', 's3:DeleteObject','arn:aws:s3:::data-lake/old.parquet',
                  context={'aws:MultiFactorAuthPresent': 'true'}),
]

print(f"{'Principal':<20} {'Action':<20} {'Resource (short)':<30} Result")
print('-' * 100)
for req in test_requests:
    result = evaluate_policies(req, all_policies)
    res_short = req.resource.split(':::')[-1]
    mfa = ' [MFA]' if req.context.get('aws:MultiFactorAuthPresent') == 'true' else ''
    print(f"{req.principal:<20} {req.action:<20} {res_short:<30} {result}{mfa}")
```

## 3.4  Cross-Cloud IAM Patterns

| Pattern | How it works | Best for |
|---------|-------------|----------|
| **Centralised IdP** | Azure AD or Okta federates to AWS IAM Identity Center and GCP Cloud Identity | Enterprises with existing IdP investment |
| **Workload Identity Federation** | Workload's native JWT exchanged for cloud credentials via OIDC | CI/CD pipelines, Kubernetes workloads |
| **Cross-account roles** | AWS role trusts GCP service account; GCP SA impersonates AWS role | AWS↔GCP direct data access |
| **Service mesh mTLS** | Mutual TLS between services; identity is the certificate | Microservice-to-microservice within a mesh |

**Golden rule for machine identities:** Never store a long-lived secret. If you cannot avoid it, rotate it every 24 hours and alert on any access older than the rotation window.

```python
# Workload Identity Federation simulator
# Models token exchange flow: native JWT → cloud STS → short-lived credentials

import hashlib
import time
from dataclasses import dataclass
from typing import Dict, Optional


@dataclass
class OIDCToken:
    """Simplified OIDC JWT claims."""
    issuer: str          # e.g. 'https://token.actions.githubusercontent.com'
    subject: str         # e.g. 'repo:acme/backend:ref:refs/heads/main'
    audience: str        # e.g. 'https://iam.googleapis.com/...'
    issued_at: float
    expiry: float

    def is_valid(self) -> bool:
        return time.time() < self.expiry


@dataclass
class CloudCredentials:
    provider: str
    access_key_id: str
    secret: str          # short-lived; in real life this is ephemeral
    expiry: float
    principal_arn: str

    def is_valid(self) -> bool:
        return time.time() < self.expiry


class WorkloadIdentityPool:
    """
    Simulates GCP Workload Identity Pool or AWS IAM OIDC provider.
    Validates an OIDC token and issues short-lived cloud credentials.
    """
    def __init__(self, provider: str, trusted_issuers: Dict[str, str]):
        self.provider = provider
        # issuer_url → mapped_role/service_account
        self.trusted_issuers = trusted_issuers
        self.ttl_seconds = 3600  # 1-hour credentials

    def exchange(self, token: OIDCToken) -> Optional[CloudCredentials]:
        if not token.is_valid():
            print(f'  [STS] REJECTED: token expired')
            return None

        mapped_principal = self.trusted_issuers.get(token.issuer)
        if not mapped_principal:
            print(f'  [STS] REJECTED: issuer not trusted: {token.issuer}')
            return None

        # Derive a deterministic-but-fake access key for simulation
        raw = f'{token.subject}:{time.time():.0f}'
        key_id = hashlib.sha256(raw.encode()).hexdigest()[:16].upper()
        secret = hashlib.sha256((raw + 'secret').encode()).hexdigest()[:32]

        creds = CloudCredentials(
            provider=self.provider,
            access_key_id=key_id,
            secret=secret,
            expiry=time.time() + self.ttl_seconds,
            principal_arn=mapped_principal,
        )
        print(f'  [STS] ISSUED credentials for principal: {mapped_principal}')
        return creds


# ── Simulate a GitHub Actions workflow authenticating to AWS and GCP ──────────
github_token = OIDCToken(
    issuer='https://token.actions.githubusercontent.com',
    subject='repo:acme/backend:ref:refs/heads/main',
    audience='sts.amazonaws.com',
    issued_at=time.time(),
    expiry=time.time() + 300,  # 5-minute OIDC token
)

aws_pool = WorkloadIdentityPool(
    provider='aws',
    trusted_issuers={
        'https://token.actions.githubusercontent.com': 'arn:aws:iam::123456789:role/github-ci-role',
    },
)

gcp_pool = WorkloadIdentityPool(
    provider='gcp',
    trusted_issuers={
        'https://token.actions.githubusercontent.com': 'projects/acme/serviceAccounts/github-ci@acme.iam.gserviceaccount.com',
    },
)

print('=== Workload Identity Federation: GitHub Actions → AWS + GCP ===')
print(f'GitHub OIDC token issuer : {github_token.issuer}')
print(f'GitHub OIDC token subject: {github_token.subject}')
print()

print('Authenticating to AWS...')
aws_creds = aws_pool.exchange(github_token)
if aws_creds:
    print(f'  Access key (first 8): {aws_creds.access_key_id[:8]}...')
    print(f'  Expires in: {int(aws_creds.expiry - time.time())} seconds')

print()
print('Authenticating to GCP...')
gcp_creds = gcp_pool.exchange(github_token)
if gcp_creds:
    print(f'  Access key (first 8): {gcp_creds.access_key_id[:8]}...')
    print(f'  Principal: {gcp_creds.principal_arn}')

print()
print('Testing expired token...')
expired_token = OIDCToken(
    issuer='https://token.actions.githubusercontent.com',
    subject='repo:acme/backend:ref:refs/heads/main',
    audience='sts.amazonaws.com',
    issued_at=time.time() - 400,
    expiry=time.time() - 100,  # already expired
)
aws_pool.exchange(expired_token)
```

## Summary

1. **Federate, don't duplicate** — A centralised IdP (Azure AD, Okta) is always cheaper to manage than parallel user directories on each cloud.
2. **Workload Identity Federation** eliminates long-lived secrets from CI/CD — every cloud supports OIDC token exchange from GitHub Actions, GitLab, Kubernetes, and other issuers.
3. **Explicit Deny wins** — AWS, GCP, and Azure all evaluate policies with the same hierarchy: explicit deny beats explicit allow beats implicit deny.
4. **Least privilege is a process** — Set boundaries tightly at provisioning time; review and shrink permissions quarterly using Access Analyzer (AWS) or Policy Insights (GCP).

**Next →** [`04_data_portability`](04_data_portability.ipynb) — open formats and cloud-agnostic storage.
