## 5.1  Fork & Pull Request Model

```bash
# 1. Fork on GitHub, then clone
git clone git@github.com:you/repo.git
git remote add upstream git@github.com:original/repo.git

# 2. Create feature branch
git switch -c feature/add-neovim-course

# 3. Make commits
git add -p                    # interactive staging (recommended)
git commit -m 'feat: add neovim course ENV102'

# 4. Keep in sync
git fetch upstream
git rebase upstream/main

# 5. Push and open PR
git push origin feature/add-neovim-course
gh pr create --title 'Add ENV102 Neovim course' \
             --body 'Closes #123'
```

## 5.2  GitHub CLI (gh)

```bash
gh repo clone user/repo
gh pr create
gh pr list
gh pr checkout 42
gh pr merge --squash
gh issue create
gh issue list
gh workflow run
gh release create v1.0.0
```

## 5.3  Commit Message Convention

```
<type>: <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
```
feat: add ENV102 Neovim mastery course
fix: correct broken TOC references in semester 5
docs: update CALIBRE_MAPPING with new books
chore: run generate_all_notebooks for expansion
```

```python
import subprocess

def git(cmd):
    r = subprocess.run(f'git {cmd}', shell=True, capture_output=True, text=True)
    return r.stdout.strip()

print('=== Repository Info ===')
print('Remotes:')
print(git('remote -v'))
print()
print('Recent commits (conventional format):')
print(git('log --oneline -8'))
```

## GitHub CLI and PR Workflow

```bash
# GitHub CLI (gh)
gh auth login              # authenticate
gh repo create my-project --public  # create repo
gh pr create --fill        # PR from current branch (fill from commits)
gh pr list                 # list open PRs
gh pr checkout 123         # check out PR locally
gh pr review 123 --approve
gh pr merge 123 --squash   # squash and merge

# CI/CD with GitHub Actions
# .github/workflows/test.yml
# name: Test
# on: [push, pull_request]
# jobs:
#   test:
#     runs-on: ubuntu-latest
#     steps:
#       - uses: actions/checkout@v4
#       - uses: astral-sh/setup-uv@v2
#       - run: uv sync && uv run pytest
```

```python
actions_workflow = '''
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ["3.11", "3.12"]

    steps:
      - uses: actions/checkout@v4

      - name: Set up uv
        uses: astral-sh/setup-uv@v2
        with:
          python-version: ${{ matrix.python-version }}

      - name: Install dependencies
        run: uv sync --dev

      - name: Lint
        run: uv run ruff check . && uv run ruff format --check .

      - name: Type check
        run: uv run mypy .

      - name: Test
        run: uv run pytest --cov=src --cov-report=xml

      - uses: codecov/codecov-action@v3
'''
print(actions_workflow)
```

```python
# Verification helpers — exercise checker

def assert_equal(got, want, label=""):
    status = "✓" if got == want else f"✗ got={got!r} want={want!r}"
    print(f"  {label or 'test'}: {status}")

def assert_approx(got, want, tol=1e-6, label=""):
    diff = abs(got - want)
    status = "✓" if diff < tol else f"✗ diff={diff:.2e}"
    print(f"  {label or 'test'}: {status}")

def run_tests(fn, cases):
    """cases: list of (args_tuple, expected)"""
    passed = sum(1 for args, want in cases
                 if (result := fn(*args) if isinstance(args, tuple) else fn(args)) == want
                 or (print(f"  FAIL {fn.__name__}{args}: got {result!r}, want {want!r}"), False))
    print(f"{fn.__name__}: {passed}/{len(cases)} passed {'✓' if passed==len(cases) else '✗'}")

# ─── Example: test your implementations below ────────────────────────────────

def flatten(nested):
    """Flatten arbitrarily nested lists."""
    result = []
    for item in nested:
        if isinstance(item, list):
            result.extend(flatten(item))
        else:
            result.append(item)
    return result

run_tests(flatten, [
    (([1,[2,[3,[4]]],5],),  [1,2,3,4,5]),
    (([],),                  []),
    (([[1],[2],[3]],),        [1,2,3]),
])

def chunk(lst, size):
    """Split list into chunks of given size."""
    return [lst[i:i+size] for i in range(0, len(lst), size)]

run_tests(chunk, [
    (([1,2,3,4,5], 2),  [[1,2],[3,4],[5]]),
    (([1,2,3], 3),       [[1,2,3]]),
    (([], 2),            []),
])
```

```python
# Advanced: Git object model — blob, tree, commit simulation
import hashlib, json
from dataclasses import dataclass, field
from typing import Optional

def sha1(content: str) -> str:
    return hashlib.sha1(content.encode()).hexdigest()[:8]

@dataclass
class GitObject:
    kind:    str   # blob | tree | commit
    content: str
    hash:    str = field(init=False)
    def __post_init__(self):
        self.hash = sha1(f"{self.kind}:{self.content}")

class GitRepo:
    def __init__(self):
        self.objects: dict[str, GitObject] = {}
        self.refs:    dict[str, str] = {"HEAD": None}
        self.index:   dict[str, str] = {}   # filename → blob hash

    def hash_object(self, kind: str, content: str) -> str:
        obj = GitObject(kind=kind, content=content)
        self.objects[obj.hash] = obj
        return obj.hash

    def add(self, filename: str, content: str) -> str:
        blob_hash = self.hash_object("blob", content)
        self.index[filename] = blob_hash
        print(f"  staged {filename} → blob {blob_hash}")
        return blob_hash

    def commit(self, message: str, author: str = "dev") -> str:
        tree_content = json.dumps(dict(sorted(self.index.items())))
        tree_hash    = self.hash_object("tree", tree_content)
        parent       = self.refs.get("HEAD")
        commit_data  = json.dumps({"tree": tree_hash, "parent": parent,
                                   "author": author, "message": message})
        commit_hash  = self.hash_object("commit", commit_data)
        self.refs["HEAD"] = commit_hash
        print(f"  commit {commit_hash}  '{message}'  tree={tree_hash}")
        return commit_hash

    def log(self):
        cur = self.refs.get("HEAD")
        while cur:
            obj = self.objects.get(cur)
            if not obj: break
            data = json.loads(obj.content)
            print(f'  commit {cur}  "{data['message']}"  tree={data['tree']}')
            cur  = data.get("parent")

repo = GitRepo()
print("=== git add ===")
repo.add("README.md",  "# My Project")
repo.add("main.py",    "print('hello')")
print("=== git commit ===")
repo.commit("Initial commit")

repo.add("main.py", "print('hello world')")   # modify file
repo.commit("Update main.py")

print("=== git log ===")
repo.log()
print(f"=== Objects stored: {len(repo.objects)} ===")
for h, obj in list(repo.objects.items())[:6]:
    print(f"  {h}  {obj.kind}")
```

## Git Quick Reference

```bash
# Daily workflow
git status -s                     # compact status
git diff --staged                 # review before commit
git commit -m "type: description" # conventional commit
git push -u origin feature/name   # first push

# Branching
git switch -c feature/name        # create + switch
git switch main && git pull        # update main
git rebase main                    # rebase feature onto main
git merge --no-ff feature/name     # merge with history

# Undoing things (safe)
git restore <file>                 # discard working tree changes
git restore --staged <file>        # unstage
git revert HEAD                    # new commit that undoes last

# History inspection
git log --oneline --graph --all    # visual branch graph
git log -p --follow -- <file>      # file history with diffs
git blame -L 10,20 <file>          # who changed lines 10-20
git bisect start                   # binary search for bug commit

# Stash
git stash push -m "WIP: feature"
git stash list
git stash pop                      # restore latest

# Remote
git remote -v
git fetch --prune                  # remove stale remote branches
git pull --rebase origin main      # rebase instead of merge pull
```

**Commit types (Conventional Commits):**
`feat` · `fix` · `refactor` · `docs` · `test` · `chore` · `perf` · `ci`

## Failure Modes

- **Secrets in CI**: never hardcode secrets in workflow YAML — use `secrets.MY_SECRET` and add in repo Settings
- **Action pinning**: `uses: actions/checkout@v4` is floating — pin to SHA for security: `uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683`

## Exercises
1. Set up a CI pipeline that runs tests on PR and deploys on main merge
2. Add a release workflow: tag → build → publish to PyPI using trusted publishing
3. Configure branch protection: require 2 reviews, passing CI, no direct push to main

## Key Takeaways

- **`gh pr create --fill`**: GitHub CLI creates a PR from the current branch, filling title and body from commit messages — eliminates switching to the browser.
- **`git add -p` (interactive staging)**: stages individual hunks from changed files — the best practice for creating atomic, focused commits from multi-purpose edits.
- **GitHub Actions `on: pull_request`**: triggers CI on every PR — the standard way to enforce tests, lint, and type checking before merge.
- **`actions/checkout@SHA` (pinned)**: pin action versions to a full SHA not a tag — tags are mutable and floating tags are a supply-chain attack vector.
- **`secrets.MY_SECRET`**: access repository secrets in workflow YAML — never hardcode tokens or credentials in workflow files committed to the repo.

## 🧠 Quick Quiz

**Q1.** GitHub Flow is a branching strategy where:
- A) Releases use long-lived release branches
- B) **`main` is always deployable; features are done in short-lived branches merged via PRs** ✓
- C) Developers commit directly to `main`
- D) Hotfixes use a dedicated `hotfix/` branch hierarchy

**Q2.** A GitHub Actions workflow is triggered by:
- A) Only `git push`
- B) **Events: push, pull_request, schedule, workflow_dispatch, and many others** ✓
- C) Only manual triggers
- D) Merges to `main` only

**Q3.** Branch protection rules enforce:
- A) Code style conventions
- B) **Required reviews, status checks, and restrictions before merging** ✓
- C) Commit message formats
- D) Maximum PR size

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
