# Reach the WAN via default gateway

PatchLab missions teach that same-subnet ping is not enough — off-subnet traffic needs a correct default gateway and a host address that actually sits on the LAN.

## Ticket

`SERVER-01` is mis-addressed (`10.10.99.10/16`) with a bogus gateway. LAN is `10.10.10.0/24`, gateway `10.10.10.1`, ISP peer `203.0.113.1`.

## Tasks

1. Set `ip=10.10.10.10`, `prefix=24`, `gateway=10.10.10.1` in `host.env`
2. Record `LAN_OK` and `WAN_OK` in `reachability.txt`

Tip codes: `BROKEN_ADDRESS`, `MASK_TRAP`, `WRONG_GATEWAY`.
