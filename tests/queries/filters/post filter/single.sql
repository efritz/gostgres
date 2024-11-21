SELECT
    film_id,
    title,
    length
FROM film
WHERE length > 120
ORDER BY
    rating,
    title
LIMIT 5;
