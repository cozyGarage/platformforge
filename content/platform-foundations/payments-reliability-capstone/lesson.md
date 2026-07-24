# Capstone — stabilize payments reliability

Checkout error rates spiked after a release. Topology defaults are unsafe, settlements drifted from payments, and there is no threshold alert. Restore the intended state, prove the ledger is clean, and leave evidence.

Treat this as a production incident:

1. Observe Compose, Kubernetes, metrics, and the ledger before editing
2. Make the smallest safe repairs
3. Reconcile data, not just config
4. Add an alert that would have paged earlier
5. Document root cause, mitigation, and follow-up, then commit

This capstone combines container networking, Kubernetes capacity, SQL reconciliation, observability, and incident communication.
