## 2.1  Core Commands

```bash
git branch                    # list local branches
git branch -a                 # list all (including remote)
git branch feature/new-thing  # create branch
git checkout feature/new-thing # switch
git switch -c feature/new      # create + switch (modern)
git branch -d branch-name      # delete merged
git branch -D branch-name      # force delete
```

## 2.2  Branching Models

### GitHub Flow (recommended for most projects)
```
main  ←─────────────────────────────────
           ↑ PR merge
feature  ──┘
```
1. Branch from `main`
2. Make commits
3. Open PR
4. Merge (squash merge preferred)

### Git Flow (complex projects)
```
main    ──●──────────────────────────────●──
           ↑ release                      ↑
develop ──●──────────────────────────────●──
           ↑                     ↑
feature ───┘        hotfix ──────┘
```

## 2.3  Useful Branch Commands

```bash
# See merged branches (safe to delete)
git branch --merged main | grep -v main

# Delete all merged feature branches
git branch --merged main | grep feature/ | xargs git branch -d

# See which branch contains a commit
git branch --contains <sha>

# Stash work-in-progress
git stash push -m "WIP: feature X"
git stash list
git stash pop
```

```python
import subprocess

def git(cmd):
    r = subprocess.run(f'git {cmd}', shell=True, capture_output=True, text=True)
    return r.stdout.strip()

print('=== Branch Status ===')
print(git('branch -v'))
print()
print('=== Recent Commits ===')
print(git('log --oneline --graph --decorate -10'))
```

## Branching Strategies

| Strategy | Main branches | When to use |
|----------|--------------|-------------|
| **GitHub Flow** | main + feature branches | Continuous deployment |
| **Git Flow** | main + develop + release + hotfix | Scheduled releases |
| **Trunk-based** | main only + short-lived feature flags | Large teams, CI/CD |

```bash
# GitHub Flow (most common)
git switch -c feature/user-auth  # create feature branch
git push -u origin feature/user-auth
# ... work, commit ...
gh pr create                     # open PR
# ... review, merge to main ...
git switch main && git pull
git branch -d feature/user-auth  # clean up
```

```python
branching_comparison = '''
GitHub Flow (recommended for most teams):
  + Simple: one protected branch (main)
  + Fast: feature → PR → merge to main → deploy
  - Requires feature flags for in-progress work
  - Hard if supporting multiple versions

Git Flow (complex projects with releases):
  + Clear release management
  + Separate develop buffer
  - Complex merge hell
  - Slow: features → develop → release → main

Trunk-Based (Google, Facebook scale):
  + Fastest integration
  + No merge conflicts (integrate many times per day)
  - Requires sophisticated feature flag infrastructure
  - Needs excellent test coverage

Choose: GitHub Flow for most projects.
         Git Flow only if you ship installable software with versioned releases.
         Trunk-based for large teams with mature CI/CD.
'''
print(branching_comparison)
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

- **Long-lived feature branches**: diverge from main, merge conflicts accumulate — rebase weekly or use trunk-based dev
- **Merging main into feature repeatedly**: creates noisy merge commits — use `git rebase origin/main` for clean linear history

## Exercises
1. Practice the GitHub Flow: create feature branch, make 3 commits, squash to 1, open PR, merge
2. Set up branch protection rules: require PR reviews, passing CI, no force push to main
3. Resolve a merge conflict deliberately: create conflicting changes on 2 branches, merge and resolve

## Key Takeaways

- **GitHub Flow**: `main` is always deployable; features live in short-lived branches merged via PRs — the standard for continuous deployment teams.
- **`git switch -c feature/name`**: the modern command to create and switch to a branch simultaneously — prefer over the older `git checkout -b`.
- **Trunk-based development (TBD)**: all developers commit to `main` at least daily using feature flags for incomplete work — reduces merge conflicts and speeds up CI.
- **`git rebase origin/main`**: replays your feature branch commits on top of the latest main — keeps history linear and avoids noisy merge commits on long-lived branches.
- **Branch protection**: enforce PR reviews and passing CI before merging to `main` via GitHub repo settings — prevents broken code from reaching the main branch.

## 🧠 Quick Quiz

**Q1.** Trunk-based development (TBD) recommends:
- A) Long-lived feature branches
- B) **All developers commit to main/trunk frequently (at least daily)** ✓
- C) One release branch per version
- D) GitFlow with develop and release branches

**Q2.** Feature flags in trunk-based development allow:
- A) Skipping code review
- B) **Merging incomplete features without exposing them to users** ✓
- C) Merging without CI
- D) Automatic versioning

**Q3.** `git cherry-pick <commit>` applies:
- A) All commits from a branch
- B) The oldest commit in a branch
- C) **A specific commit from another branch to the current branch** ✓
- D) The latest changes without committing

<details><summary>Answers</summary>1-B, 2-B, 3-C</details>
