# Connect and push to a remote

Local commits are private until you publish them. Remotes are how platform teams share history — Boot.dev **Learn Git** covers `origin`, push, and upstream tracking.

## Why remotes matter

CI, code review, and collaboration all depend on a shared remote. Before you open a PR, `main` (or your feature branch) must exist on `origin` with tracking set so later `git pull` / `git push` know where to go.

## Starting state

- A local repo in `/workspace` with one commit on `main` (`README.md`)
- An empty bare remote already exists at `/tmp/platform.git`
- No remotes are configured yet

## Tasks

1. Add the bare repo as remote `origin`
2. Push `main` with upstream tracking (`-u`)
3. Confirm `origin` has `refs/heads/main` and local `main` tracks `origin/main`

## Useful commands

```sh
git remote add origin /tmp/platform.git
git push -u origin main
git remote -v
git ls-remote --heads origin
```
