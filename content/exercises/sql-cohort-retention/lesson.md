# Cohort retention in SQL

Averages hide churn. A cohort table answers the real question: of the people who first showed up in January, how many were still active in February, March, and so on? The shape is always the same: **assign each user a cohort** (the month of their first event), **reduce activity to distinct (user, month) pairs**, then **count users per cohort and month offset** against the cohort size.

The classic mistakes are counting events instead of users (a busy user inflates every month) and integer division (`2 / 3` is `0` in SQLite). The tests check both. The reading *Cohort Analysis: Retention, Churn and LTV* and *Window Functions* cover the technique.

## Task

Edit `/workspace/retention.sql` so one query returns `cohort_month`, `months_since`, `users`, `retention`, ordered by `cohort_month`, `months_since`.

The data is in `/workspace/events.db`, table `events(user_id, event_date)`.
