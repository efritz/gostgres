SELECT
    films.film_id,
    films.title,
    films.category_name
FROM (
    SELECT
        f.film_id,
        f.title,
        c.name AS category_name
    FROM film f
    JOIN film_category fc ON fc.film_id = f.film_id
    JOIN category c ON c.category_id = fc.category_id
) films
ORDER BY films.film_id
LIMIT 5;
