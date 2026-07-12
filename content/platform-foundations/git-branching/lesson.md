# Branch and merge safely

Feature branches are how platform teams ship changes without destabilizing `main`. This lab practices the merge workflow Boot.dev teaches in **Learn Git**.

## Starting state

- `main` has `app.env` with `version=1.0.0`
- You are on `feature/timeout` with unstaged changes adding `TIMEOUT=30`

## Tasks

1. Commit the change on `feature/timeout`
2. Switch to `main` and merge the feature branch
3. Confirm `app.env` on `main` contains `TIMEOUT=30`
