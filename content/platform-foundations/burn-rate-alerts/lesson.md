# Design Multi-Window Burn Alerts

Burn-rate alerts detect when the error budget is draining too quickly — fast windows page, slow windows ticket.

## Why it matters

A single threshold either pages too often or wakes nobody until the month is already ruined. Multi-window burn separates acute from sustained risk.

## Ticket focus

1. `PaymentsBurnFast` (1h → page)
2. `PaymentsBurnSlow` (6h → ticket)
3. On-call note with rollback first

## Tip codes

- `BURN_FAST_SLOW` — short window pages; longer window tickets
- `PAGE_VS_TICKET` — match severity to window
- `ROLLBACK_FIRST` — fast burn mitigation starts with rollback
