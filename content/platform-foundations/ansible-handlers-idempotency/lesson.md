# Handlers and Idempotency

Handlers fire when a task reports **changed**. Idempotent playbooks can run repeatedly with the same desired end state.

## Why it matters

Noise from unnecessary restarts hides real failures. Clean change signals make GitOps and CI reviews trustworthy.

## Ticket focus

Document `/workspace/docs/IDEMPOTENCY.md` with headings:

1. **Changed** — notify/handler when a task reports change
2. **Skipped** — unchanged/already present means no handler
3. **Check mode** — `--check` / dry-run before production

## Tip codes

- `HANDLER_ALWAYS` — handlers run only when a task reports changed
