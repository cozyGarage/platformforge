# Establish a Debian platform baseline

Golden images start as written standards: SSH hardened, packages pinned, evidence saved. This lab captures that baseline without needing a real VM.

## Starting state

- `etc/ssh/sshd_config` allows root passwords
- `baseline/packages.txt` lists unpinned package names

## Tasks

1. Set `PermitRootLogin no` and `PasswordAuthentication no`
2. Pin packages to `nginx=1.22.1` and `postgresql=16.4`
3. Write `baseline/STANDARD.md` documenting key-only SSH and pinned packages
