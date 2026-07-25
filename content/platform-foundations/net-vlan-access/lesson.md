# Fix a VLAN access mismatch

In PatchLab, amber LEDs mean the cable is plugged but L2 adjacency fails. The same idea shows up in tickets: physical path looks fine, VLAN identity does not.

## Ticket

`SERVER-07` expects **VLAN 20**. It is currently on `Gi1/0/2` (**VLAN 10**). Keep documentation label `A-07`.

## Tasks

1. Attach `SERVER-07` to `Gi1/0/7` (VLAN 20)
2. Clear the old attachment on `Gi1/0/2`
3. Keep `panel_label: A-07`
4. Record `PATH_UP VLAN_20` in `/workspace/net/STATUS.txt`

Tip code to remember: `VLAN_MISMATCH`.
