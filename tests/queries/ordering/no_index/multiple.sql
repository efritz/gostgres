SELECT
    film_id,
    title,
    rating,
    rental_rate
FROM film
ORDER BY
    rating,
    rental_rate
LIMIT 5;
