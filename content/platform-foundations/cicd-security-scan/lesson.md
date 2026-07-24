# Add a supply chain security gate

Shipping without a vulnerability scan — or with secrets in git — is how incidents start. Boot.dev **Learn CI/CD** pairs quality gates with supply-chain checks before deploy.

## Starting state

- `.github/workflows/ci.yaml` has `test`, `lint`, and `deploy` (deploy already needs test/lint)
- `secrets.env` contains a live-looking API key that must never ship
- `.gitignore` does not ignore `secrets.env`

## Tasks

1. Add a `security` job with `runs-on: ubuntu-latest` and a step that runs `trivy` or `npm audit`
2. Update `deploy` so it `needs` the `security` job
3. Delete `secrets.env` and add `secrets.env` to `.gitignore`

## Example security job

```yaml
security:
  runs-on: ubuntu-latest
  steps:
    - run: trivy fs .
deploy:
  runs-on: ubuntu-latest
  needs: [test, lint, security]
  steps:
    - run: echo deploy
```
