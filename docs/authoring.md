# Authoring PlatformForge labs

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
