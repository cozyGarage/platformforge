# Plan External Secrets Sync

External Secrets Operator syncs credentials from a vault/backend into Kubernetes Secrets without storing plaintext in Git.

## Why it matters

Committed tokens leak forever in history. A SecretStore + ExternalSecret keeps Git as policy and the backend as truth.

## Ticket focus

1. SecretStore scope and auth
2. ExternalSecret target + refresh
3. Policy: Allowed vs Forbidden

## Tip codes

- `STORE_SCOPE` — namespaced store unless a shared cluster store is justified
- `STORE_AUTH` — IRSA / Kubernetes SA / AppRole — no long-lived tokens in Git
- `ESO_TARGET` — writes a native Secret via `spec.target.name`
- `NO_PLAINTEXT` — never commit `password:` / `token:` values
- `GIT_NO_SECRET` — Git holds references and sync policy, not credentials
