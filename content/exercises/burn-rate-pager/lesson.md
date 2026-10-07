# Page on error-budget burn rate

An alert on "error rate above 1%" ignores your SLO: 1% is a disaster for a 99.99% service and nothing for a 95% one. **Burn rate** fixes that by measuring how fast you are spending the error budget *relative to the SLO*. A burn rate of 1 uses exactly the whole budget by the end of the period; 14.4 uses it in two days.

A single window is a poor pager. A long window alone keeps firing after the incident ends; a short window alone fires on every blip. Requiring **both** to be hot pages on real, ongoing burn only. This is the multi-window rule from the Google SRE workbook, and the reading *SRE Fundamentals* and *Alerting and On-Call* cover why.

## Tasks

1. `budget_minutes` and `burn_rate`: the arithmetic everything else builds on.
2. `should_page`: two windows, both at or above the threshold.
3. `hours_to_exhaustion`: the number an on-call engineer actually wants in the page.

Run all tests with `cd /workspace && python -m unittest discover -s tests`.
