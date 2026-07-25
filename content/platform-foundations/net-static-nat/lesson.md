# Publish a host with static NAT

PatchLab’s static NAT mission publishes an internal host for inbound WAN reachability. Platform teams do the same for demos and partner callbacks — map one inside address to one outside address, then permit that path.

## Ticket

Publish `SERVER-01` `10.10.10.10` as `203.0.113.10`. Allow WAN → that public IP.

## Tasks

1. Configure `nat.yaml` with `mode: static`, inside/outside addresses
2. Put a permit for `203.0.113.10` above the broad deny in `acl.yaml`
3. Record `NAT_OK INBOUND_OK` in `STATUS.txt`

Tip codes: `NAT_MISSING`, `IMPLICIT_DENY`.
