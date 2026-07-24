# Design a private artifact bucket

CI artifacts belong in a private, versioned bucket with lifecycle expiry and TLS-only access — never a public ACL.

## Tasks

1. Update `/workspace/s3/bucket.json`:
   - `Bucket`: `payments-artifacts`
   - `PublicAccessBlock`: `true`
   - `Versioning`: `Enabled`
   - `LifecycleDays`: `30`
2. Write `/workspace/s3/bucket-policy.json` denying `s3:GetObject` when `aws:SecureTransport` is false
