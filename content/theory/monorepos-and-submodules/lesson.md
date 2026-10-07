## 6.1  Git Submodules

```bash
# Add submodule
git submodule add git@github.com:user/lib.git libs/lib

# Clone repo with submodules
git clone --recurse-submodules git@github.com:user/repo.git

# Update all submodules
git submodule update --remote --merge

# Remove submodule
git submodule deinit libs/lib
git rm libs/lib
```

## 6.2  Sparse Checkout (Large Repos)

```bash
git clone --filter=blob:none --sparse git@github.com:big/repo.git
git sparse-checkout set path/to/subdir another/path
```

## 6.3  Useful Git Aliases

```bash
# ~/.gitconfig
[alias]
    st = status --short
    lg = log --oneline --graph --decorate
    undo = reset --soft HEAD~1
    wip = !git add -A && git commit -m 'WIP'
    unwip = reset HEAD~1 --mixed
    aliases = config --get-regexp alias
    recent = branch --sort=-committerdate --format='%(refname:short)' -10
```

```python
import subprocess

def git(cmd):
    r = subprocess.run(f'git {cmd}', shell=True, capture_output=True, text=True)
    return r.stdout.strip()

print('=== Git Aliases Configured ===')
aliases = git('config --get-regexp alias')
if aliases:
    for line in aliases.splitlines():
        parts = line.split(None, 1)
        if len(parts) == 2:
            print(f'  {parts[0].replace("alias.",""):15} = {parts[1]}')
else:
    print('  (no aliases configured)')
    print()
    print('Add to ~/.gitconfig:')
    print('  [alias]')
    print('  lg = log --oneline --graph --decorate')
```

## Monorepos and Submodules

```bash
# Git submodules: embed a repo within a repo
git submodule add https://github.com/org/shared-lib libs/shared
git submodule update --init --recursive   # clone all submodules
git submodule foreach git pull origin main  # update all

# Alternatives to submodules:
# - git subtree (merge history, no .gitmodules)
# - npm/pip packages (separate releases)
# - monorepo with workspace tools

# Monorepo with pnpm workspaces
# pnpm-workspace.yaml:
# packages:
#   - 'packages/*'
#   - 'apps/*'
pnpm -r build           # build all packages
pnpm -r test            # test all packages
pnpm --filter=my-app dev  # run only my-app
```

```python
monorepo_structure = '''
Monorepo structure (Python + TypeScript):

my-monorepo/
├── packages/
│   ├── shared-utils/
│   │   └── pyproject.toml      # shared Python utilities
│   └── ml-models/
│       └── pyproject.toml      # ML model package
├── apps/
│   ├── web-frontend/
│   │   └── package.json        # React app
│   └── api-server/
│       └── pyproject.toml      # FastAPI app
├── uv.toml                     # workspace config
├── pnpm-workspace.yaml         # JS workspace
└── Makefile                    # cross-language commands

# uv workspaces (uv.toml)
[tool.uv.workspace]
members = ["packages/*", "apps/api-server"]

# Makefile
test:
    uv run --package shared-utils pytest packages/shared-utils/tests
    uv run --package api-server pytest apps/api-server/tests
    pnpm --filter web-frontend test
'''
print(monorepo_structure)
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

- **Submodule detached HEAD**: after `submodule update`, submodules are in detached HEAD — `cd libs/shared && git checkout main` to track branch
- **Forgetting `--init --recursive`**: cloned repo without submodules — run `git submodule update --init --recursive`

## Exercises
1. Add a git submodule for a shared library, update it to a specific tag
2. Set up a uv workspace with 2 packages that depend on each other
3. Configure Turborepo or Nx for monorepo build caching and dependency graph

## Key Takeaways

- **`git submodule add <url> path`**: embeds another repository at a specific commit inside your repo — the submodule pointer is tracked as a special entry in the parent repo.
- **`git submodule update --init --recursive`**: the command to run after cloning a repo with submodules — without it, submodule directories are empty.
- **Submodule detached HEAD**: after `git submodule update`, submodules are pinned to a commit not a branch — `cd submodule && git checkout main` to track the branch actively.
- **`git subtree`**: alternative to submodules that merges the external repo's history — simpler for contributors but creates noisier history.
- **Monorepo benefits**: a single repo for multiple packages enables atomic cross-package commits, shared CI, and unified code search — tools like Turborepo and Nx add caching.

## 🧠 Quick Quiz

**Q1.** Git submodules allow:
- A) Splitting a repo into multiple branches
- B) **Including another git repository at a specific commit inside your repo** ✓
- C) Sharing branches between repos
- D) Syncing histories between repos

**Q2.** A monorepo contains:
- A) A single application only
- B) **Multiple projects in one repository, sharing tooling and version history** ✓
- C) Only shared libraries
- D) Multiple branches per project

**Q3.** `git submodule update --init --recursive` is needed because:
- A) Submodules auto-update on clone
- B) **Submodule directories are empty after a fresh clone; this command populates them** ✓
- C) It upgrades submodule branches
- D) It re-initialises the parent repo

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
