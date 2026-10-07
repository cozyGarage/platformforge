## 4.1  What Are Git Hooks?

Scripts in `.git/hooks/` that run at specific git events:

| Hook | When | Common Use |
|------|------|------------|
| `pre-commit` | Before commit | Lint, format, tests |
| `commit-msg` | After msg entered | Validate format |
| `pre-push` | Before push | Run full test suite |
| `post-merge` | After pull/merge | `uv sync` |

## 4.2  Example pre-commit Hook

```bash
#!/usr/bin/env bash
# .git/hooks/pre-commit
set -e

echo "Running pre-commit checks..."

# Run ruff linter
if command -v ruff &>/dev/null; then
  ruff check . || { echo "Lint failed"; exit 1; }
fi

# Check for debug prints
if git diff --cached | grep -E '^\+.*print\(' | grep -v '#'; then
  echo "Warning: print() statements in commit"
  # uncomment to make this blocking:
  # exit 1
fi

echo "Pre-commit OK"
```

## 4.3  pre-commit Framework (Recommended)

```bash
pip install pre-commit
```

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.3.0
    hooks:
      - id: ruff
      - id: ruff-format
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-json
      - id: check-yaml
```

```bash
pre-commit install    # install hooks
pre-commit run --all-files  # run manually
```

```python
import os, pathlib

hook_dir = pathlib.Path('.git/hooks')
if hook_dir.exists():
    hooks = list(hook_dir.iterdir())
    print('=== Git Hooks ===')
    for h in sorted(hooks):
        executable = os.access(h, os.X_OK)
        status = '✓ active' if executable and not h.name.endswith('.sample') else '○ sample'
        print(f'  {status}  {h.name}')
else:
    print('Not in a git repo with hooks directory')
```

## Git Hooks for Automation

```bash
# Hooks live in .git/hooks/ (local, not tracked)
# Use pre-commit framework for shareable hooks

# .pre-commit-config.yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
        args: [--fix]
      - id: ruff-format

  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-merge-conflict
      - id: detect-private-key    # prevent accidental secret commits

# Install hooks
pre-commit install            # installs .git/hooks/pre-commit
pre-commit run --all-files    # run on all files (once)
```

```python
precommit_script = '''#!/usr/bin/env bash
# .git/hooks/pre-commit (manual, without pre-commit framework)
set -euo pipefail

echo "Running pre-commit checks..."

# 1. Check for debug statements
if git diff --cached --name-only | xargs grep -l "import pdb\|breakpoint()\|debugger;" 2>/dev/null; then
    echo "ERROR: Debug statements found — remove before committing"
    exit 1
fi

# 2. Check for large files (> 1MB)
for file in $(git diff --cached --name-only); do
    if [[ -f "$file" ]] && [[ $(stat -f%z "$file" 2>/dev/null || stat -c%s "$file") -gt 1048576 ]]; then
        echo "ERROR: Large file detected: $file (> 1MB)"
        exit 1
    fi
done

# 3. Run tests for changed Python files
changed_py=$(git diff --cached --name-only --diff-filter=ACM | grep '\\.py$' || true)
if [[ -n "$changed_py" ]]; then
    python -m pytest --tb=short -q tests/ || { echo "Tests failed"; exit 1; }
fi

echo "Pre-commit checks passed."
'''
print(precommit_script)
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

- **Hooks not executable**: hooks must be `chmod +x .git/hooks/pre-commit`; the pre-commit framework handles this
- **Bypassing hooks**: `git commit --no-verify` skips all hooks — should be exceptional; CI should enforce the same checks independently

## Exercises
1. Set up `pre-commit` with ruff, mypy, and detect-private-key
2. Write a commit-msg hook that enforces conventional commit format: `type(scope): description`
3. Add a post-merge hook that automatically runs `uv sync` when `pyproject.toml` changes

## Key Takeaways

- **`.git/hooks/` scripts**: executable scripts triggered by git events — `pre-commit`, `commit-msg`, `pre-push`, `post-merge` are the most commonly automated hooks.
- **`pre-commit` framework**: manages shareable hooks via `.pre-commit-config.yaml` — installs hooks from public repos (ruff, mypy, detect-private-key) with version pinning.
- **`pre-commit install`**: writes the hook into `.git/hooks/pre-commit` — must be run once per clone; add to `make install` so new contributors get it automatically.
- **`commit-msg` hook**: validates that commit messages match a pattern (e.g. Conventional Commits `type: description`) — enforces team standards before the commit is created.
- **`git commit --no-verify`**: bypasses all hooks — should be exceptional; duplicate the same checks in CI so bypassed hooks don't let bad code reach main.

## 🧠 Quick Quiz

**Q1.** `pre-commit` git hooks run:
- A) After code is pushed to remote
- B) **Before the commit is created — can reject the commit if they exit non-zero** ✓
- C) Before merging branches
- D) During rebasing

**Q2.** `commit-msg` hook validates:
- A) File content
- B) **The commit message format** ✓
- C) Test results
- D) Branch name

**Q3.** `pre-push` hook is useful for:
- A) Formatting code before commits
- B) **Running tests before allowing a push to remote** ✓
- C) Signing commits
- D) Creating release tags

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
