# Linux shell and file basics

> **Attribution:** Scenario adapted from [90DaysOfDevOps](https://github.com/MichaelCade/90DaysOfDevOps) Linux fundamentals (Michael Cade, CC BY-NC-SA 4.0). Rewritten as an interactive PlatformForge exercise.

You are onboarding to a platform team. A ticket is waiting in `/workspace/inbox/TICKET.md`. The shell starts in `/workspace`. The `logs/` directory is not there yet — creating it is part of the exercise.

## What you are practicing

1. **Navigation.** `pwd` prints the current directory. `ls` lists it. `ls inbox` lists a subdirectory without leaving `/workspace`. A path that starts with `/` is absolute and ignores your current directory. `notes.txt` and `./notes.txt` are relative to wherever you are.
2. **Creating a file with redirection.** `>` sends command output into a file and creates the file if needed. Write `/workspace/notes.txt` so it contains the word `platformforge`.
3. **A new directory and a copy.** `mkdir /workspace/logs` creates the backup directory. `cp` places the same note at `/workspace/logs/notes.bak`.

## Done when

- `/workspace/notes.txt` contains `platformforge`
- `/workspace/logs` exists because you created it
- `/workspace/logs/notes.bak` contains `platformforge`

The validator checks those files, not the exact commands. `printf`, `echo`, and copying a file you already wrote are all fine.

## Why this matters

Release automation and incident response both happen in a shell. Paths, redirection, and copies are the moves underneath later labs on permissions and log pipelines.

## Source

- Original curriculum: https://github.com/MichaelCade/90DaysOfDevOps/tree/main/2022/Days
- Modifications: condensed narrative, added deterministic validators and isolated lab environment
