# Capstone — design a platform landing zone

Before a FinTech payments platform lands in a new environment, the baseline must already answer least privilege, encryption in transit, network tiers, Kubernetes capacity, and rollback. This lab is a local simulation — no cloud account required.

Deliver the pack:

1. Replace wildcard IAM with a scoped artifacts reader
2. Deny S3 access that is not over TLS (`aws:SecureTransport`)
3. Plan public and private VPC subnets
4. Baseline the payments Deployment (3 replicas, non-root)
5. Write a runbook + short ADR and commit the evidence

This capstone stitches AWS mental models, Kubernetes reliability, and platform documentation into one release-ready folder.
