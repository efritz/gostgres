-- SKIP
-- Doesn't work without explicit GROUP BY

SELECT
    min(rental_rate) AS lowest_rate,
    max(rental_rate) AS highest_rate
FROM film;
