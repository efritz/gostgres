SELECT
    film_id,
    title,
    rental_rate
FROM film
ORDER BY film_id
LIMIT 5
OFFSET 10;
