-- SKIP
-- Not implemented in parser

SELECT
    *,
    rental_rate * rental_duration AS total_rental_cost
FROM film
ORDER BY film_id
LIMIT 5;
