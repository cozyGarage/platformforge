# Recover without rewriting history

The release branch is missing `app.env`. The file existed in an earlier commit, and the branch is already shared, so rewriting history is unsafe.

Inspect the repository, restore the tracked file, and create a new commit that records the recovery. This leaves an auditable trail and avoids disrupting collaborators.

## Operational habit

Before applying a Git recovery command, identify whether the unwanted state is in the working tree, index, local commits, or published history. Choose the narrowest reversible operation.
