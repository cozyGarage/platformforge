# Order firewall ACLs correctly

PatchLab’s firewall missions punish putting a broad permit above a specific deny. Real ACLs behave the same way: **first match wins**.

## Ticket

Keep `10.10.10.10` able to reach WAN. Block only `10.10.10.20/32`.

## Tasks

1. First ACL rule: deny `10.10.10.20/32`
2. Later rule: permit `10.10.10.0/24`
3. Record `DENY_20` and `ALLOW_10` in `probe.txt`

Tip code: `ACL_ORDER`.
