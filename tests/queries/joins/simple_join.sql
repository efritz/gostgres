SELECT
    f.film_id,
    f.title,
    c.name AS category_name
FROM film f
JOIN film_category fc ON f.film_id = fc.film_id
JOIN category c ON fc.category_id = c.category_id
ORDER BY f.film_id
LIMIT 5;
