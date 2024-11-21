SELECT
    film_id,
    title,
    rating,
    rental_rate
FROM film
ORDER BY
    rating ASC,
    rental_rate DESC
LIMIT 5;
