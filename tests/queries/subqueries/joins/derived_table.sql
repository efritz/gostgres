-- SKIP
-- Can't resolve f.film_id

SELECT
    f.film_id,
    f.title,
    c.category_name
FROM film f
JOIN (
    SELECT
        fc.film_id,
        c.name AS category_name
    FROM film_category fc
    JOIN category c ON c.category_id = fc.category_id
) c ON c.film_id = f.film_id
ORDER BY f.film_id
LIMIT 5;
