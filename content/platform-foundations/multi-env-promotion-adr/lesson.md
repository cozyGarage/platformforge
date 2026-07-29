# ADR — Multi-Env Promotion Path

Promotions should be Git commits of Helm values/overlays, reconciled by GitOps, with secrets outside Git.

## Why it matters

Laptop `kubectl` per env creates snowflakes. Digest-pinned values plus ExternalSecret keep audits and rollbacks honest.

## Ticket focus

1. ADR 0002 choosing GitOps/Helm values promotion
2. Dev → Stage → Prod checklist

## Tip codes

- `PROMOTE_GIT` — promotions are commits, not laptop applies
- `DIGEST_UP` — same digest upward
- `ENV_LADDER` — dev → stage → prod
