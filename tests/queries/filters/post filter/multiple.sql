SELECT
    film_id,
    title,
    length
FROM film
WHERE
    length > 120 AND
    rental_rate > 2.99
ORDER BY
    rating,
    title
LIMIT 5;
