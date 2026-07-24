# Trace a service connection path

> **Attribution:** Scenario adapted from [90DaysOfDevOps](https://github.com/MichaelCade/90DaysOfDevOps) networking topics (Michael Cade, CC BY-NC-SA 4.0). Rewritten as an interactive PlatformForge exercise.

Before packets move, names resolve and ports open. Platform engineers debug connectivity as a path: DNS/hosts → TCP → HTTP.

## Tasks

1. Write `/workspace/net/path.env`:
   - `HOST=payments.local`
   - `PORT=8080`
   - `HEALTH_URL=http://payments.local:8080/health`
2. Write `/workspace/net/NOTES.md` explaining hosts/DNS resolution before the TCP port connect

## Source

- Original curriculum: https://github.com/MichaelCade/90DaysOfDevOps
- Modifications: condensed into a connection-path documentation drill with validators
