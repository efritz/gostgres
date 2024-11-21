SELECT
    film_id,
    title,
    length
FROM film
WHERE
    title < 'C' AND
    length > 120
ORDER BY
    rating,
    title
LIMIT 5;
