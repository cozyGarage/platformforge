# Plan a Proxmox homelab platform

Before you click Create VM, platform work starts on paper: node capacity, workload placement, and network segmentation.

## Tasks

1. Edit `/workspace/plan/inventory.yaml`:
   - Node `pve01` with `cpu: 8`, `memory_gb: 64`, `storage_gb: 1000`
   - VM `db01` role `postgres` on `pve01`
   - VM `kafka01` role `kafka` on `pve01`
2. Write `/workspace/plan/network.md` describing a management VLAN and a storage VLAN

## Example inventory sketch

```yaml
nodes:
  - name: pve01
    cpu: 8
    memory_gb: 64
    storage_gb: 1000
vms:
  - name: db01
    node: pve01
    role: postgres
  - name: kafka01
    node: pve01
    role: kafka
```
