SELECT
    film_id,
    title,
    rental_rate,
    rental_duration
FROM film
ORDER BY rental_rate * rental_duration
LIMIT 5;