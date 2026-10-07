# Authoring PlatformForge labs and readings

Each lab is a directory containing `lab.yaml` and `lesson.md`. Validate authored content with:

```bash
make lint
```

Manifests follow `schemas/lab.schema.json`. Use versioned, minimal images and deterministic setup. Prefer typed checks:

- `command`: a bounded command exits successfully;
- `file`: a path exists and optionally contains a fixed value;
- `process`: a matching process exists;
- `port`: a TCP listener exists;
- `http`: an endpoint responds;
- `docker` and `kubernetes`: domain-specific declarative commands.

Never place credentials in a lab, mount the host Docker socket, request privileged containers, or rely on hidden learner command history. Validate the desired end state so learners can discover alternative correct solutions.

### k3d runtime addons

For `runtime.type: k3d`, optional `runtime.addons` let the host prepare cluster capabilities **before** the learner shell starts (still no privileged lab containers, still no docker.sock):

| Addon | Host behavior |
|-------|----------------|
| `kyverno` | `kubectl apply` Kyverno install; wait for admission controller |
| `gateway` | Install Gateway API CRDs (`standard-install`) |
| `gateway-controller` | Gateway API CRDs + Envoy Gateway controller |
| `otel` | Apply embedded Jaeger + OTel Collector manifests; wait Ready |
| `multi-cluster` | Create a second k3d cluster; merge kubeconfig with `east` / `west` contexts |
| `etcd-snapshot` | `docker exec` into the k3d server node, take an etcd/k3s snapshot, mount `snapshot.db` into the lab |
| `registry` | `k3d cluster create --registry-create`; export `REGISTRY` / `REGISTRY_HOST` |

Optional `runtime.hostBuild` (requires `registry`) copies the lab build context to the host, runs `docker build` / `docker push`, and writes `/workspace/IMAGE` for deploy tasks. Learners author the Dockerfile; they never receive a Docker socket.

Every task needs a clear objective, progressive hints, actionable validation names, deterministic reset through setup commands, and a lesson explaining the operational reason behind the exercise.

### Pedagogy notes (shared with PatchLab)

Prefer **action → truth → tip** over long pre-reads:

1. Seed a broken or incomplete workspace in `setup`
2. Put the symptom in a short ticket file when the lab is incident-shaped
3. Name checks after the failure mode learners should recognize (`VLAN mismatch`, `Deny appears before permit`)
4. Keep hints progressive: symptom decode → concrete fix → exact expected artifact

Networking ticket labs under `net-*` follow this pattern intentionally.

Runtime UX (do not re-implement in each lab):

- Failed validates increment per-task counters; after every 2 failures a **ghost hint** auto-reveals
- Passing validate writes a **debrief score** (correctness / speed / cleanliness → stars)
- `prerequisites` are enforced on lab start; path modules may also declare `unlock.completedFromModule` + `unlock.count` for sandbox gates

Adapted material belongs only in `content/90days`. Include the source URL, original author, modification notes, an `attribution` manifest field, and CC BY-NC-SA 4.0 licensing.

## Readings (theory units)

A reading is a lesson with no lab: a directory under `content/` holding `reading.yaml` and `lesson.md`.

```yaml
version: 1
id: helm-charts-and-releases      # lowercase-hyphen; must not collide with any lab id (they share one progress table)
title: Helm Chart Structure and Release Management
summary: One or two sentences, at least 10 characters.
estimatedMinutes: 30
prerequisites: []
source: where it was adapted from (optional)
quiz:                              # optional self-check; answer is the zero-based index
  - q: In Helm value precedence (lowest to highest), which order is correct?
    options: ["`--set` < `-f override.yaml` < `values.yaml`", "`values.yaml` < `-f override.yaml` < `--set`"]
    answer: 1
    explain: optional
```

- List a reading in a path module with `readings: [id]` (rendered before the module's labs).
- `lesson.md` supports GitHub-style Markdown and `mermaid` fenced diagrams, rendered in the browser; the library loads only on pages that contain one.
- The quiz key is served to the browser: readings are self-study, not assessment. "Mark as read" records completion.
- Convert a notebook with `scripts/nb2reading.py` (markdown cells kept, code cells fenced, outputs dropped, `## Vault Insights` and `## After This Notebook` cells skipped; `--mcq` lifts selected questions from an MCQ notebook). Review the result: it is a starting point, not a finished lesson.
- `go test ./internal/content` checks that every reading loads, its quiz is valid, its id is unique against labs, and every path reference resolves.
