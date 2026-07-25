# Author a Terraform root module

IaC starts with a root module someone else can open and understand. Pin the Terraform version, name the app in `locals`, and export that name as an output.

## Tasks

1. Create `/workspace/infra/main.tf`
2. Set `required_version = ">= 1.5.0"`
3. Define `locals.app_name = "payments"`
4. Output `app_name` from `local.app_name`
5. Record `PLAN_READY`

Tip code: `MISSING_VERSION`.
