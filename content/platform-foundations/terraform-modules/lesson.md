# Split a Terraform network module

Root modules grow messy fast. Pull the VPC CIDR into `modules/network` with a variable and output, then call it from the root with a concrete CIDR.

## Layout

```text
infra/
  main.tf
  modules/network/
    main.tf
    variables.tf
    outputs.tf
```

## Tasks

1. Add `variable "vpc_cidr"` and `output "vpc_cidr"`
2. Use `var.vpc_cidr` inside the module
3. Call `module "network"` with `source = "./modules/network"` and `vpc_cidr = "10.0.0.0/16"`
4. Record `MODULE_OK`

Tip code: `MISSING_MODULE_SOURCE`.
