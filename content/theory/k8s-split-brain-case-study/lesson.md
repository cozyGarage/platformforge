A custom controller `inventory-reconciler` running 3 replicas with Kubernetes lease-based leader election experienced a split-brain on 2026-04-02: a brief 70-second control-plane partition let two pods both believe they were the leader, and each one began applying a different reconciliation plan. The result was 142 inventory rows written twice and 18 written in conflicting orders.

This lesson reconstructs the failure mode, builds a corrected leader election with **fencing tokens**, simulates the recovery, and produces the runbook for the next time it happens (it will).

## What "split-brain" actually means here

```
            kube-apiserver (lease holder of truth)
                 |       |       |
              partition  ✓       ✓
                 |       |       |
                 v       v       v
              pod-A    pod-B    pod-C
              (was leader, can't renew lease)   (sees lease expired, becomes leader)

         partition heals
                 |
                 v
   pod-A still thinks it's leader  +  pod-C is the real leader
   => two writers, no fencing => corruption
```

The mistake is assuming "I hold the lease" is *sufficient* to act. It isn't: by the time the action reaches the downstream system, you may no longer be the leader. The fix is **fencing tokens**: a monotonically increasing number issued with each lease grant, included in every downstream write, and validated by the downstream system.

## Part 1 — Bad election: lease without fencing

```python
import time, random
from dataclasses import dataclass, field

@dataclass
class Lease:
    holder: str | None = None
    expiry: float = 0.0
    epoch: int = 0   # monotonically increasing per acquisition

class Apiserver:
    def __init__(self, ttl=10.0):
        self.lease = Lease()
        self.ttl = ttl
        self.now = 0.0
    def tick(self, dt): self.now += dt
    def try_acquire(self, holder):
        if self.now >= self.lease.expiry or self.lease.holder == holder:
            new_epoch = self.lease.epoch + (0 if self.lease.holder == holder else 1)
            self.lease = Lease(holder, self.now + self.ttl, new_epoch)
            return True, self.lease.epoch
        return False, self.lease.epoch
    def renew(self, holder):
        if self.lease.holder == holder and self.now < self.lease.expiry:
            self.lease.expiry = self.now + self.ttl
            return True
        return False

class UnsafePod:
    def __init__(self, name, api): self.name = name; self.api = api; self.is_leader = False
    def step(self):
        if not self.is_leader:
            ok, _ = self.api.try_acquire(self.name)
            self.is_leader = ok
        else:
            ok = self.api.renew(self.name)
            if not ok:  # we *think* we're still leader but lost the lease
                pass     # bug: we don't notice until next renew, keep writing

api = Apiserver(ttl=10.0)
pods = [UnsafePod(f"pod-{x}", api) for x in "ABC"]

# Phase 1: normal -- pod-A becomes leader
api.tick(0.1); pods[0].step()
print(f"t={api.now:5.1f} leader={api.lease.holder} epoch={api.lease.epoch}")
# Phase 2: partition -- pod-A can't see apiserver for 12s, lease expires
api.tick(12.0)
# Phase 3: pod-C acquires
pods[2].step()
print(f"t={api.now:5.1f} leader={api.lease.holder} epoch={api.lease.epoch}  pod-A still thinks={pods[0].is_leader}")
```

**Where the bug is.** `pod-A` keeps `is_leader=True` across the partition. Any write it issues during the next few hundred ms after the partition heals (before its next failed renew) lands in the downstream system *without* being labelled as stale. The downstream has no way to reject it.

## Part 2 — Fenced writes: the downstream is the arbiter

```python
class InventoryDB:
    def __init__(self): self.rows: dict[str, tuple] = {}; self.highest_epoch = 0
    def write(self, key, value, epoch):
        # Fencing rule: never accept a write with epoch < highest seen.
        if epoch < self.highest_epoch:
            return ("REJECT_STALE_EPOCH", self.highest_epoch)
        self.highest_epoch = max(self.highest_epoch, epoch)
        self.rows[key] = (value, epoch)
        return ("OK", epoch)

class FencedPod:
    def __init__(self, name, api, db):
        self.name = name; self.api = api; self.db = db
        self.is_leader = False; self.my_epoch = 0
    def step(self):
        if not self.is_leader:
            ok, ep = self.api.try_acquire(self.name)
            self.is_leader, self.my_epoch = ok, ep
        else:
            if not self.api.renew(self.name):
                # demoted; remember our epoch but stop writing
                self.is_leader = False
    def reconcile(self, key, value):
        if not self.is_leader: return ("not-leader", None)
        return self.db.write(key, value, self.my_epoch)

api = Apiserver(ttl=10.0); db = InventoryDB()
A = FencedPod("pod-A", api, db); C = FencedPod("pod-C", api, db)

api.tick(0.1); A.step()
print("A acquires:", A.reconcile("widget-1", "qty=10"))

# Partition: A loses connectivity (cannot renew); time advances 12s
api.tick(12.0)
# pod-C acquires (new epoch)
C.step()
print("C acquires:", C.reconcile("widget-1", "qty=20"))

# Partition heals — A tries to write before it notices it's been demoted
print("A writes stale:", A.reconcile("widget-1", "qty=99"))
```

**What just happened.** `pod-A`'s reconcile carries `epoch=1`; the DB has already seen `epoch=2` from `pod-C` and rejects the write as `REJECT_STALE_EPOCH`. The split-brain still *occurred* — two pods believed they were leader — but the downstream made it *harmless*.

This is the same pattern as ZooKeeper's `zxid`, etcd's revision number, and (canonically) the fencing token in Kleppmann's *Designing Data-Intensive Applications* §8.4.

## Part 3 — Idempotency keys for the writes that do land

```python
class IdempotentDB(InventoryDB):
    def __init__(self):
        super().__init__()
        self.seen_keys: set[tuple] = set()    # (idempotency_key,)
    def write(self, key, value, epoch, idem=None):
        if idem is not None and idem in self.seen_keys:
            return ("OK_DUP", epoch)
        result = super().write(key, value, epoch)
        if result[0] == "OK" and idem is not None:
            self.seen_keys.add(idem)
        return result

db2 = IdempotentDB()
print(db2.write("a", 1, epoch=1, idem="r-001"))
print(db2.write("a", 1, epoch=1, idem="r-001"))   # retry
print(db2.write("a", 2, epoch=2, idem="r-002"))
```

**Why both.** Fencing solves *split-brain* (two pods think they're leader). Idempotency keys solve *at-least-once retries* (the network ate the response and the controller retried). They are independent layers — production controllers need both.

## Part 4 — Liveness vs safety: the CAP-flavoured trade-off

| Knob | Smaller | Larger | Trade |
|---|---|---|---|
| Lease TTL | faster failover (good liveness) | fewer mistaken takeovers (good safety) | partition longer than TTL ⇒ takeover; shorter ⇒ stuck |
| Renew interval | tolerate jitter | fewer apiserver writes | too large ⇒ unnecessary failovers |
| Lease leniency | strict (lose lease at TTL) | grace period | grace ⇒ split-brain window |
| Downstream check | fencing token | nothing | fencing makes safety apiserver-independent |

**Rule:** *safety* belongs to the downstream system (fencing). *Liveness* belongs to the apiserver (lease TTL). Don't try to make one do the other's job.

## Part 5 — Reconciler invariants worth testing

```python
# These are the invariants that should hold across any partition / failover.

INVARIANTS = []

def inv(label):
    def deco(fn):
        INVARIANTS.append((label, fn)); return fn
    return deco

@inv("at most one epoch active at a time per row")
def i1(db, history):
    per_row = {}
    for key, val, epoch, result in history:
        if result[0] not in ("OK", "OK_DUP"): continue
        if key in per_row and per_row[key] != epoch:
            return False, f"{key} had epoch {per_row[key]} then {epoch}"
        per_row[key] = epoch
    return True, ""

@inv("rejected writes never appear in db")
def i2(db, history):
    rejected = [(k,v,e) for k,v,e,r in history if r[0].startswith("REJECT")]
    for k, v, _ in rejected:
        if k in db.rows and db.rows[k][0] == v:
            return False, f"rejected value for {k} ended up in db"
    return True, ""

@inv("highest epoch is monotonic")
def i3(db, history):
    last = 0
    for _,_,e,r in history:
        if r[0] in ("OK","OK_DUP"):
            if e < last: return False, f"epoch went {last} -> {e}"
            last = max(last, e)
    return True, ""

# Replay a richer scenario through the fenced design
api2 = Apiserver(ttl=10.0); db3 = InventoryDB()
A2 = FencedPod("pod-A", api2, db3); C2 = FencedPod("pod-C", api2, db3)
history = []
def w(pod, k, v):
    r = pod.reconcile(k, v)
    history.append((k, v, pod.my_epoch, r))
api2.tick(0.1); A2.step()
w(A2, "x", 1); w(A2, "y", 2)
api2.tick(12.0); C2.step()
w(C2, "x", 10)
w(A2, "x", 99)  # stale
w(A2, "y", 3)   # also stale

for label, fn in INVARIANTS:
    ok, msg = fn(db3, history)
    print(f"  [{'PASS' if ok else 'FAIL'}] {label}{'  -- '+msg if msg else ''}")
print("\nfinal db:", db3.rows)
```

## Part 6 — Runbook

| Alert | First 5 min | Next 30 min | Postmortem must answer |
|---|---|---|---|
| `leader_election_flap > 3/min` | Page on-call; confirm apiserver health, not controller bug | Diff lease TTL vs observed partition lengths | Is TTL too aggressive for this control plane? |
| `REJECT_STALE_EPOCH from downstream` | Confirm no data corruption; identify stale leader | Audit logs for what stale leader tried to write | Did the stale leader self-terminate fast enough? |
| `two distinct leader_id observed in same minute` | Treat as in-progress split-brain; verify fencing is on | Check downstream rejection rate | What was the partition root cause (apiserver/etcd)? |
| `controller_restart_loop` | Roll back recent controller image | Bisect to faulty commit | Did the new code respect fencing? |

**Rules of thumb**
- Lease TTL should be **at least 3× the worst observed network jitter** to the apiserver. Sub-second leases are how you create split-brain, not avoid it.
- The downstream system, not the controller, decides what is safe. If the downstream can't fence, you do not have safety — period.
- `kubectl get lease -A` is your first command; `kubectl logs --previous` is your second.

## Reflection

**Takeaways**
- Lease-based leader election is a *liveness* mechanism. Safety needs a separate fencing layer.
- Fencing tokens are cheap to add and convert split-brain from "corruption" to "rejected writes".
- Idempotency keys and fencing tokens solve different problems; you need both for at-least-once delivery + safety.
- Test the invariants ("at most one active epoch per row"), not just the happy path.
- Lease TTLs are a network-jitter problem, not a UX problem — pick them from p99 latency, not from intuition.


**Going further.** Replace the in-process apiserver with a real `k8s.io/client-go` lease, wire fencing into a PostgreSQL backend using an advisory lock keyed by epoch, and add a chaos-test that artificially extends the partition window and asserts zero downstream commits from the stale leader.
