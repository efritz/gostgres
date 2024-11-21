SELECT
    f.id AS film_id,
    f.movie_name AS title,
    f.cat AS category_name
FROM (
    SELECT
        f.film_id,
        f.title,
        c.name
    FROM film f
    JOIN film_category fc ON fc.film_id = f.film_id
    JOIN category c ON c.category_id = fc.category_id
) f(id, movie_name, cat)
ORDER BY f.id
LIMIT 5;
