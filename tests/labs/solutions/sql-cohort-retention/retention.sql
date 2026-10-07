WITH first_month AS (
  SELECT user_id, strftime('%Y-%m', MIN(event_date)) AS cohort_month FROM events GROUP BY user_id
),
activity AS (
  SELECT DISTINCT user_id, strftime('%Y-%m', event_date) AS active_month FROM events
),
sizes AS (
  SELECT cohort_month, COUNT(*) AS size FROM first_month GROUP BY cohort_month
)
SELECT f.cohort_month,
       (CAST(substr(a.active_month, 1, 4) AS INTEGER) - CAST(substr(f.cohort_month, 1, 4) AS INTEGER)) * 12
         + CAST(substr(a.active_month, 6, 2) AS INTEGER) - CAST(substr(f.cohort_month, 6, 2) AS INTEGER) AS months_since,
       COUNT(DISTINCT a.user_id) AS users,
       ROUND(1.0 * COUNT(DISTINCT a.user_id) / s.size, 2) AS retention
FROM first_month f
JOIN activity a ON a.user_id = f.user_id
JOIN sizes s ON s.cohort_month = f.cohort_month
GROUP BY f.cohort_month, months_since
ORDER BY f.cohort_month, months_since;
