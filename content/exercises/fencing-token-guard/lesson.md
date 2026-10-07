# Fence a stale leader

Leader election tells a process it is the leader. It cannot tell a *paused* process that it stopped being the leader while it slept. A controller that wakes from a long GC pause and keeps writing will overwrite the work of the leader that replaced it: that is split-brain.

A **fencing token** closes the gap. Every lease grant carries a number higher than all earlier ones; the *storage* remembers the highest token it has seen and refuses anything lower. The safety lives in the resource being protected, because it is the only party that sees both leaders. Idempotency keys handle the other half: a retried request must not apply twice.

The full story is in the reading *Case Study: Split-Brain Recovery in a Leader-Elected Controller*.

## Tasks

1. `LeaseStore.acquire` issues strictly increasing tokens.
2. `FencedStorage.write` rejects stale tokens, changes nothing when it does, and ignores retried request ids.

Run all tests with `cd /workspace && python -m unittest discover -s tests`.
