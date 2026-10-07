## 1.1  Everything is a Hash

Git stores 4 types of objects, all content-addressed by SHA-1:

| Object | Contains |
|--------|----------|
| **blob** | File contents |
| **tree** | Directory (list of blobs + trees) |
| **commit** | Tree + parent + metadata |
| **tag** | Named commit pointer |

```bash
# Inspect objects
git cat-file -t HEAD          # type
git cat-file -p HEAD          # content of commit
git cat-file -p HEAD^{tree}   # content of tree

# See the object database
ls .git/objects/
```

## 1.2  Refs

Refs are human-readable pointers to SHA-1 hashes:
```
.git/
├── refs/
│   ├── heads/main     ← branch pointer
│   ├── heads/feature  ← another branch
│   └── remotes/origin/main
├── HEAD               ← current branch
└── ORIG_HEAD          ← pre-merge/rebase state
```

## 1.3  The Three Trees

```
Working Directory → Index (Staging) → Repository
      git add →          git commit →
```

- **Working directory**: what you see in the filesystem
- **Index/Stage**: snapshot prepared for next commit
- **HEAD**: last commit

```python
import subprocess

def git(cmd):
    r = subprocess.run(f'git {cmd}', shell=True, capture_output=True, text=True)
    return r.stdout.strip()

print('=== Git Object Model Exploration ===')
print('HEAD commit:', git('rev-parse HEAD')[:16], '...')
print('HEAD type:  ', git('cat-file -t HEAD'))
print()
print('Last commit:')
print(git('log --oneline -5'))
print()
print('Object database size:')
print(git('count-objects -vH'))
```

## Git Object Database

Git stores 4 types of objects, all content-addressed (SHA-1/SHA-256 hash):

| Object | Content |
|--------|---------|
| **blob** | File contents |
| **tree** | Directory listing (blob + tree refs) |
| **commit** | Snapshot: root tree + parent(s) + message + author |
| **tag** | Annotated tag: points to commit + message |

```bash
git cat-file -t <hash>   # type of object
git cat-file -p <hash>   # print object content
git log --oneline        # commit hashes + messages
git show HEAD:src/main.py  # view file at HEAD
git ls-tree HEAD         # list root tree
```

```python
import subprocess, pathlib, tempfile, os

# Show git object model on current repo
try:
    # Get HEAD commit hash
    head = subprocess.check_output(['git', 'rev-parse', 'HEAD'], text=True).strip()
    print(f'HEAD commit: {head[:12]}')

    # Show commit object
    commit = subprocess.check_output(['git', 'cat-file', '-p', 'HEAD'], text=True)
    print('Commit object:')
    for line in commit.splitlines()[:6]:
        print(f'  {line}')

    # Get tree from commit
    tree_hash = [l for l in commit.splitlines() if l.startswith('tree')][0].split()[1]
    print(f'\nRoot tree: {tree_hash[:12]}')
    tree = subprocess.check_output(['git', 'cat-file', '-p', tree_hash], text=True)
    for line in tree.splitlines()[:5]:
        print(f'  {line}')
except subprocess.CalledProcessError:
    print('Not in a git repo or git not available')
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

- **Detached HEAD**: checkout a commit directly — you're not on a branch, new commits will be orphaned; always `git checkout -b new-branch` from detached state
- **Shallow clones**: `git clone --depth 1` creates shallow repo — `git log`, `git blame`, `git rebase` all break; use `git fetch --unshallow` to fix

## Exercises
1. Walk the object graph: `git log --oneline | head -3` → pick a commit → `git cat-file -p <hash>` → find the tree → read a blob
2. Manually create a commit: `git hash-object -w file.txt` → `git write-tree` → `git commit-tree`
3. Explain in your own words why git history is immutable (hint: hash of commit includes parent hash)

## Key Takeaways

- **Four git object types**: blob (file contents), tree (directory), commit (snapshot + metadata), tag (named reference) — all content-addressed by SHA-1 hash.
- **`git cat-file -t <hash>`**: reveals the type of any git object; `git cat-file -p <hash>` prints its contents — essential for understanding what git actually stores.
- **`git reflog`**: records every HEAD movement — the safety net for recovering from accidental resets, rebases, and branch deletions for 90 days.
- **Detached HEAD**: occurs when you checkout a commit directly (not a branch) — any new commits will be orphaned unless you immediately `git checkout -b new-branch`.
- **Shallow clones (`--depth 1`)**: truncate history to save bandwidth, but break `git log`, `git blame`, and `git rebase` — avoid in development repos.

## 🧠 Quick Quiz

**Q1.** `git rebase` vs `git merge`: rebase:
- A) Is safer for shared branches
- B) **Replays commits on top of another branch, creating a linear history** ✓
- C) Creates a merge commit
- D) Is always faster

**Q2.** `git stash` saves:
- A) Committed changes to remote
- B) **Uncommitted working directory changes to a temporary stack** ✓
- C) The entire repository state
- D) Only staged changes

**Q3.** `git bisect` helps:
- A) Split a branch
- B) **Find the commit that introduced a bug via binary search** ✓
- C) Compare two branches
- D) Squash commits

<details><summary>Answers</summary>1-B, 2-B, 3-B</details>
