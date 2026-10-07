## 3.1  Rebase vs Merge

```
# Merge: preserves history, adds merge commit
A──B──C──D (main)
         ↑ merge commit
   E──F──G (feature)

# Rebase: clean linear history
A──B──C──E'──F'──G' (feature rebased onto main)
```

## 3.2  Interactive Rebase

```bash
git rebase -i HEAD~5    # rework last 5 commits

# pick   → keep commit
# reword → change message
# edit   → amend commit
# squash → merge into previous
# fixup  → squash, discard message
# drop   → remove commit
```

## 3.3  Common History Operations

```bash
# Amend last commit
git commit --amend

# Squash last 3 commits
git rebase -i HEAD~3  # mark 2nd and 3rd as 'squash'

# Fix commit message
git rebase -i HEAD~1  # mark as 'reword'

# Cherry-pick a specific commit
git cherry-pick <sha>

# Reset (CAREFUL)
git reset --soft HEAD~1   # undo commit, keep staged
git reset --mixed HEAD~1  # undo commit, keep working dir
git reset --hard HEAD~1   # undo commit AND changes

# Recover lost commits
git reflog                # shows all HEAD movements
git checkout <sha>        # restore to any point
```

```python
import subprocess

def git(cmd):
    r = subprocess.run(f'git {cmd}', shell=True, capture_output=True, text=True)
    return r.stdout.strip()

print('=== Commit Graph ===')
print(git('log --oneline --graph -10'))
print()
print('=== Reflog (last 5 entries) ===')
print(git('reflog --oneline -5'))
```

## Rebase and Interactive Rebase

```bash
# Rebase feature onto latest main
git switch feature/auth
git rebase main          # replay feature commits on top of main

# Interactive rebase: clean up commits before PR
git rebase -i HEAD~4    # edit last 4 commits
# In editor:
# pick a1b2c3d First commit
# squash b2c3d4e WIP: fix typo     ← squash into previous
# reword c3d4e5f Add validation    ← reword message
# drop d4e5f6a debug logging       ← remove this commit

# Fixup commit: add a fix to an earlier commit
git add .
git commit --fixup=HEAD~2     # create fixup commit
git rebase -i --autosquash HEAD~4  # auto-squash fixups
```

```python
rebase_workflow = '''
Rebase vs Merge comparison:

Merge (preserves history):
  A---B---C  (feature)
         \
  D---E---F---G  (main after merge)
  + Clear history of when features were merged
  - Noisy history with merge commits

Rebase (linear history):
  D---E---A\'---B\'---C\'  (rebased feature on main)
  + Clean, linear history
  + Easy to bisect bugs
  - Rewrites commits (never rebase shared branches!)

Golden rule: NEVER rebase branches that others have pulled from.
             Only rebase local branches or feature branches before merging.

Interactive rebase use cases:
  1. Squash "WIP" commits before PR
  2. Reorder commits for logical grouping
  3. Edit commit messages
  4. Split a large commit into smaller ones
  5. Remove accidental debug commits
'''
print(rebase_workflow)
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

- **Rebasing pushed commits**: rewrites SHA — anyone who pulled the old commits will have diverged history; use `git push --force-with-lease` (not `--force`) if you must
- **Losing commits during rebase**: use `git reflog` to find lost commits — they're still in the object database for 90 days

## Exercises
1. Squash 5 WIP commits into 1 meaningful commit using interactive rebase
2. Split a commit that changed 3 unrelated files: `git rebase -i HEAD~1` → `edit` → `git reset HEAD^` → stage and commit separately
3. Use `git bisect` to find which commit introduced a bug: `git bisect start`, `git bisect good v1.0`, `git bisect bad HEAD`

## Key Takeaways

- **`git rebase -i HEAD~N`**: interactive rebase lets you squash, reorder, edit, or drop the last N commits before opening a PR — the standard way to clean up WIP history.
- **`squash` vs `fixup`**: `squash` merges commits and lets you edit the combined message; `fixup` silently discards the squashed commit's message — use fixup for typo fixes.
- **Never rebase public branches**: rebase rewrites SHA hashes — anyone who pulled the old commits will have diverged history requiring a force push.
- **`git push --force-with-lease`**: safer than `--force` — fails if someone else has pushed since your last fetch, preventing accidental overwrites of others' work.
- **`git reflog` recovery**: squashed or reset commits remain in the object database for 90 days — `git reflog` shows every HEAD position, enabling recovery with `git checkout <hash>`.

## 🧠 Quick Quiz

**Q1.** Interactive rebase (`git rebase -i`) allows:
- A) Merging branches interactively
- B) **Reordering, squashing, editing, and dropping commits** ✓
- C) Resolving merge conflicts visually
- D) Staging individual lines

**Q2.** `git rebase` should not be used on:
- A) Local feature branches
- B) **Public/shared branches — it rewrites history that others may have based work on** ✓
- C) Experimental branches
- D) Branches with a single commit

**Q3.** `git reflog` is useful for:
- A) Viewing the remote repository log
- B) **Recovering lost commits after a reset or rebase** ✓
- C) Listing all branches
- D) Showing file change history

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
