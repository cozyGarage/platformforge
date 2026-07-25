# Repair a broken traceroute path

PatchLab’s traceroute mission is about reading hops and fixing the missing route — not memorizing CLI flags. When the path dies at the firewall, look for a missing more-specific route.

## Ticket

Traceroute from `SERVER-01` to `198.51.100.1` stops at the FW. Expected path:

`SERVER-01 → FW-LAN → FW-WAN/ISP-PEER → destination`

## Tasks

1. Add FW route `198.51.100.0/24` via `203.0.113.1`
2. Write the four successful hops to `trace.ok.txt`
3. Record `TRACE_COMPLETE` in `STATUS.txt`

Tip code: `NO_ROUTE`.
