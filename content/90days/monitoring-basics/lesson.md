# Build a tiny uptime check

> **Attribution:** Scenario adapted from [90DaysOfDevOps](https://github.com/MichaelCade/90DaysOfDevOps) monitoring topics (Michael Cade, CC BY-NC-SA 4.0). Rewritten as an interactive PlatformForge exercise.

Monitoring starts with a boring question: is it up? This lab turns probe status codes into a machine-readable status file.

## Data

`/workspace/monitor/probes.txt`:

```
payments 200
checkout 500
```

## Tasks

1. Write executable `/workspace/monitor/check.sh`
2. Produce `/workspace/monitor/status.env` with `payments=up` and `checkout=down`
3. Write `/workspace/monitor/ALERT.md` mentioning `checkout`

## Source

- Original curriculum: https://github.com/MichaelCade/90DaysOfDevOps
- Modifications: probe-to-status script scenario with deterministic validators
