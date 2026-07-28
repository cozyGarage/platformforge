# Write an Error Budget Policy

Error budget is the allowed unreliability. When it is gone, change the release posture — not the narrative.

## Why it matters

FinTech platforms need an explicit freeze / slow-roll / page ladder so product and platform argue from the same numbers.

## Ticket focus

1. Allowed vs actual failure math
2. Remaining budget (can be negative)
3. Freeze / Slow-roll / Page policy

## Tip codes

- `BUDGET_MATH` — remaining ≈ allowed_failure − actual_failure
- `FREEZE_GATE` — over budget pauses risky deploys
- `BURN_FAST` — fast drain pages; slow drain tickets
