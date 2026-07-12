# Filesystem discovery

Running out of disk on a node is a classic platform incident. `du` and `find` are your first tools before reaching for volume expansion.

## Tasks

1. Run `du -sh /workspace/data` and save the output to `/workspace/disk-audit.txt`
2. Find files larger than 1 MB under `/workspace/data` and save paths to `/workspace/large-files.txt`

You should discover `temp.bin` (5 MB) and `backup.bin` (2 MB) in the large-files list.
