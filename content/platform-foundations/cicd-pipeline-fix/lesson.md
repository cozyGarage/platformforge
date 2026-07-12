# Fix a broken CI pipeline

Boot.dev **Learn CI/CD** emphasizes tests and linting before deploy. Your pipeline currently jumps straight to production.

## Task

Edit `/workspace/.github/workflows/ci.yaml` to add:

1. A `test` job with `runs-on: ubuntu-latest`
2. A `lint` job with `runs-on: ubuntu-latest`
3. Keep the existing `deploy` job

Optional: use `needs: [test, lint]` on deploy so quality gates run first.
