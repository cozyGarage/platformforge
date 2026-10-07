# Linux navigation and permissions

A copied deploy script landed in `/workspace/releases/` without its executable bit. The same workspace has a short application log in `/workspace/logs/`. Fix the script's mode, then write an incident report that contains only the error lines.

The shell starts in `/workspace`. `pwd` prints that directory. `ls releases` and `ls logs` show the two places this lab uses. The paths below are absolute, so they still work after `cd`.

## Restore the deploy script

`/workspace/releases/deploy.sh` starts as mode `644` (`rw-r--r--`): readable, not executable.

Make the script executable, and leave the world-write bit off. Any of these pass:

- `chmod u+x` on the original file, which turns `644` into `744`
- `chmod 755`, `750`, or `700`

`644` still fails, because the script cannot be executed. `777` fails, because everyone can write it.

The validator reads those mode bits. It does not require one `chmod` number, and it does not look at your command history.

## Extract the incident report

`/workspace/logs/app.log` contains three lines:

```
INFO ready
ERROR disk
ERROR timeout
```

Write `/workspace/errors.txt` with exactly those two `ERROR` lines and nothing else. Order does not matter. A pipeline plus redirection is the usual approach (`grep` into the file). `awk` or `sed` also pass when the file is those two lines.

## Why this matters

Failed automation is often a path or a mode, and an incident handoff is often a filter over a noisy log. The next lab stays on pipelines. This one is only the script mode and the two-line error report.
