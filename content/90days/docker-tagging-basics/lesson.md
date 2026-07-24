# Pin and tag container images

> **Attribution:** Scenario adapted from [90DaysOfDevOps](https://github.com/MichaelCade/90DaysOfDevOps) containers topics (Michael Cade, CC BY-NC-SA 4.0). Rewritten as an interactive PlatformForge exercise.

Floating tags like `latest` make rollbacks guesswork. 90Days container days emphasize explicit image versions — PlatformForge turns that into a concrete Dockerfile fix.

## Tasks

1. Change `FROM alpine:latest` → `FROM alpine:3.21`
2. Write `payments-api:1.0.0` to `/workspace/IMAGE_TAG.txt`
3. Document in `/workspace/docs/tagging.md` that `latest` is forbidden for production

## Source

- Original curriculum: https://github.com/MichaelCade/90DaysOfDevOps
- Modifications: focused on tagging/pinning with file validators (no Docker daemon required in-lab)
