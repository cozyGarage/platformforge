# Networking sandbox — free play

PatchLab unlocks sandbox after Mission 3. Same idea here: finish the first three networking tickets, then design a small coherent LAN yourself.

## Requirements

1. An access port on **VLAN 10**
2. Host addressing `10.10.10.10/24` with gateway `10.10.10.1`
3. ACL order: deny `10.10.10.20/32` before permit `10.10.10.0/24`
4. Record `SANDBOX_OK`

There is no single “correct” switchport name — only the contracts above.
