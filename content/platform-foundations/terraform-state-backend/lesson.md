# Configure a remote state backend

Local `terraform.tfstate` does not survive a laptop swap. Platform teams store state remotely with locking so two applies cannot collide.

## Ticket values

- bucket: `payments-tf-state`
- key: `platforms/payments/terraform.tfstate`
- region: `us-east-1`
- encrypt: `true`
- dynamodb_table: `payments-tf-locks`

## Tasks

1. Add `backend "s3"` inside the `terraform` block in `main.tf`
2. Document **Backend** and **Locking** in `/workspace/docs/STATE.md`
3. Mention S3 and DynamoDB in that runbook

Tip code: `MISSING_BACKEND`.
