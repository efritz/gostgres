SELECT
    f.film_id,
    f.title,
    c.category_name
FROM film f
JOIN (
    SELECT
        fc.film_id,
        c.name
    FROM film_category fc
    JOIN category c ON c.category_id = fc.category_id
) c(id, category_name) ON c.id = f.film_id
ORDER BY f.film_id
LIMIT 5;
