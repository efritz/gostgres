SELECT
    f.film_id,
    f.title,
    c.name
FROM film f
JOIN film_category fc ON f.film_id = fc.film_id
JOIN category c ON fc.category_id = c.category_id
WHERE
    f.rental_rate > 4.50 AND
    c.name = 'Action'
ORDER BY
    rating,
    title
LIMIT 5;