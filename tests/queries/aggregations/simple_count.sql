-- SKIP
-- COUNT(*) doesn't work
-- Doesn't work without explicit GROUP BY

SELECT count(*) AS total_films
FROM film;
