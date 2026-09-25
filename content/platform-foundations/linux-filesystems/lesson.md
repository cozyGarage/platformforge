# Filesystem discovery

`/workspace/data` is filling up. Measure that directory, then list the files that are actually large. `readme.txt` is tiny on purpose, so searching by name is not enough.

## Measure the directory

`du -sh /workspace/data` prints one line: a human-readable size and the path. Save that line to `/workspace/disk-audit.txt`.

```sh
du -sh /workspace/data > /workspace/disk-audit.txt
```

## List files larger than 1 MiB

This lab runs on Alpine, where `find` is BusyBox. BusyBox size suffixes are `c` (bytes), `k` (kibibytes), and `b` (512-byte blocks). The usual GNU form fails:

```sh
find /workspace/data -size +1M
# find: invalid number '1M'
```

`+1024k` means larger than 1024 KiB (1 MiB):

```sh
find /workspace/data -type f -size +1024k > /workspace/large-files.txt
```

That list includes:

- `cache/temp.bin` (5 MiB)
- `archive/backup.bin` (2 MiB)

It does not include `readme.txt`.

## Why this matters

`du` shows which directory is heavy. `find -size` shows which files to move or delete. The check looks for both large names, rejects `readme.txt`, and expects the audit line to mention the data directory.
