# Capstone — ship a compliant data release

A vendor needs a customer extract before the payments export goes live. Raw PII is still on disk, DORA evidence is missing, the audit pack is empty, and the IAM role is `*:*`. Close every gap before the release commit.

Release checklist:

1. Mask the export so PAN/SSN/email cannot leak
2. Compute deployment frequency and change-fail rate from logs
3. Document retention, access control, and change approval
4. Replace wildcard IAM with a scoped export policy
5. Write a short release note and commit the evidence pack

This capstone combines compliance masking, DORA metrics, audit writing, and least-privilege IAM.
