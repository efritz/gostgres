SELECT
    f.film_id,
    f.title,
    c.name AS category_name,
    l.name AS language_name
FROM film f
JOIN film_category fc ON f.film_id = fc.film_id
JOIN category c ON fc.category_id = c.category_id
JOIN language l ON f.language_id = l.language_id
ORDER BY f.film_id
LIMIT 5;
