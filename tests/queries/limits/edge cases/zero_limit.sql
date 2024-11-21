-- SKIP
-- Fails

SELECT
    film_id,
    title,
    rental_rate
FROM film
ORDER BY film_id
LIMIT 0;
