-- SKIP
-- Can't resolve f.id

SELECT
    f.id AS film_id,
    f.title,
    c.cat AS category_name
FROM film f
JOIN (
    SELECT
        fc.film_id,
        c.name
    FROM film_category fc
    JOIN category c ON c.category_id = fc.category_id
) c(id, cat) ON c.id = f.film_id
ORDER BY f.id
LIMIT 5;
