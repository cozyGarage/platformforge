# Apply Kyverno Mutate Policies

Validate denies bad objects. Mutate fixes them — injecting platform labels before the Pod is persisted.

## Why it matters

FinOps and policy engines key off labels. Mutate policies make the golden path automatic instead of a wiki step.

## Ticket focus

1. ClusterPolicy `mutate-cost-center`
2. Create a bare Pod; prove `cost-center=payments` appears

## Tip codes

- `KYN_MUTATE` — mutate.patchStrategicMerge (or equivalent) for labels
- `KYN_PROVE` — read live Pod labels after create
