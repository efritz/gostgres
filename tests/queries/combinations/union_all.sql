    SELECT
        first_name,
        last_name
    FROM actor
    WHERE actor_id <= 15
UNION ALL
    SELECT
        first_name,
        last_name
    FROM actor
    WHERE
        actor_id >= 10 AND
        actor_id < 50
ORDER BY
    first_name,
    last_name;
