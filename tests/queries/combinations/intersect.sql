    SELECT
        first_name,
        last_name
    FROM actor
    WHERE actor_id <= 15
INTERSECT
    SELECT
        first_name,
        last_name
    FROM actor
    WHERE actor_id >= 10
ORDER BY
    first_name,
    last_name;