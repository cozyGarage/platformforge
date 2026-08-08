# Apply Workloads Across Two k3d Clusters

Fleet planning becomes real when the same desired state lands on more than one kube-apiserver. This lab gives you `east` and `west` contexts from two k3d clusters.

## Why it matters

Copy-paste apply without context discipline is how regions drift. Named contexts + the same manifest keep the fleet honest.

## Ticket focus

1. `kubectl --context east|west apply` payments-web
2. Wait Available on both
3. Document East / West / Waves

## Tip codes

- `FLEET_CTX` — always pass `--context` explicitly
- `FLEET_APPLY` — same YAML, two apiservers, keep wave discipline
