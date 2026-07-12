# Linux shell and file basics

> **Attribution:** Scenario adapted from [90DaysOfDevOps](https://github.com/MichaelCade/90DaysOfDevOps) Linux fundamentals (Michael Cade, CC BY-NC-SA 4.0). Rewritten as an interactive PlatformForge exercise.

Platform engineering starts in the terminal. Before containers and clusters, you need confidence with paths, files, and simple pipelines.

## Scenario

You are onboarding to a platform team. A runbook expects a notes file and a backup copy under `/workspace/logs/`.

## Tasks

Work from `/workspace`:

1. Create `notes.txt` containing the word `platformforge`
2. Ensure the `logs/` directory exists
3. Copy `notes.txt` to `logs/notes.bak`

## Why this matters

Release automation, incident response, and debugging all happen through shells. Muscle memory with basic file operations saves time when systems are down.

## Source

- Original curriculum: https://github.com/MichaelCade/90DaysOfDevOps/tree/main/2022/Days
- Modifications: condensed narrative, added deterministic validators and isolated lab environment
