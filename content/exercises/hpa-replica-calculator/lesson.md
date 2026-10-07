# Predict what the HPA will do

The Horizontal Pod Autoscaler is a control loop with one formula: `desired = ceil(current * currentMetric / targetMetric)`. Everything else is damping: a **tolerance band** so a metric at 108% of target does not cause churn, **min/max clamps**, and a **stabilisation window** so a one-minute dip in traffic does not shrink a deployment you will need back in two.

If you can compute the answer by hand, `kubectl describe hpa` stops being mysterious. The reading *Autoscaling: HPA and VPA* has the theory, including multiple metrics and why memory is a poor scaling signal.

## Tasks

1. `desired_replicas`: the formula, the tolerance band, the clamps.
2. `stabilized`: scale up immediately, scale down only to the window's highest recommendation.

Run all tests with `cd /workspace && python -m unittest discover -s tests`.
