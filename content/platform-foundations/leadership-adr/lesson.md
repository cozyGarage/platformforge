# Write an architecture decision record

Senior platform work is not only clusters — it is decisions others can trust later. An ADR freezes *why* you chose a path.

## Context notes

See `/workspace/context/notes.txt`: hostname routing, NodePort limitations, TLS at the edge.

## Tasks

1. Author `/workspace/docs/adr/0001-ingress-edge.md` with:
   - **Status** → `Accepted`
   - **Context**
   - **Decision** → choose **Ingress** (not NodePort)
   - **Consequences** → mention **TLS**
2. Add `/workspace/docs/adr/README.md` linking to that ADR
