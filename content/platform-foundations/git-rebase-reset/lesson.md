# Rebase and reset without panic

Boot.dev's **Learn Git** covers rewriting local history before you push. Platform engineers use this when a branch has noisy `wip` commits that should become one reviewable change.

## Why squash?

Reviewers prefer one logical commit per feature. Interactive rebase lets you combine commits without losing file content.

## Starting state

- Three commits on `main`: initial config, a typo fix, and retries
- `config.env` already has `retries=3`

## Tasks

1. Run `git rebase -i HEAD~2` and squash the last two commits
2. Set the final message to `feat: add retries`
3. Confirm only two commits remain and the latest message is correct
