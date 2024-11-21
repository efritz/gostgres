SELECT
    film_id,
    title,
    length
FROM film
WHERE title < 'C'
ORDER BY
    rating,
    title
LIMIT 5;
