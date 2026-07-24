# Model least-privilege IAM

Local AWS simulations teach the mental model before you touch a real account. IAM policies are JSON documents that answer: who can do which API calls on which resources.

## Tasks

1. Edit `/workspace/iam/deployer-policy.json`
2. Allow `s3:GetObject` and `s3:ListBucket` on `payments-artifacts` (+ `/*`)
3. Use Sid `ReadArtifacts` and Effect `Allow`
4. Write role name `payments-deployer` to `/workspace/iam/role.txt`

## Example statement

```json
{
  "Sid": "ReadArtifacts",
  "Effect": "Allow",
  "Action": ["s3:ListBucket", "s3:GetObject"],
  "Resource": [
    "arn:aws:s3:::payments-artifacts",
    "arn:aws:s3:::payments-artifacts/*"
  ]
}
```
