# Log pipelines and redirection

Platform on-call engineers live in log streams. Pipes and redirection turn megabytes of noise into a one-screen incident summary.

## Scenario

`/workspace/logs/app.log` contains mixed INFO, WARN, and ERROR lines from a payment service. Build a minimal triage pack for the on-call handoff.

## Tasks

1. Write only `ERROR` lines to `/workspace/error-report.txt`
2. Write the number of error lines to `/workspace/error-count.txt` (expected: `3`)

## Example commands

```sh
grep ERROR /workspace/logs/app.log > /workspace/error-report.txt
wc -l < /workspace/error-report.txt > /workspace/error-count.txt
```
