# Incident: checkout failures after release

Checkout error rates rose immediately after a release. The handoff contains a Compose topology with unsafe defaults and an under-scaled Kubernetes deployment. Restore the intended state, preserve evidence, and leave the repository clean.

Treat this as an incident:

1. Observe the supplied state before editing.
2. Make the smallest safe changes.
3. Validate each layer.
4. Record the root cause, mitigation, and a concrete follow-up.
5. Commit the recovery for auditability.

The capstone combines Linux navigation, Git recovery, container configuration, service networking, Kubernetes reliability, and incident communication.
