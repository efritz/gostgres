SELECT
    f.film_id,
    f.title,
    f.rental_rate
FROM film f
ORDER BY f.film_id
LIMIT 5;
