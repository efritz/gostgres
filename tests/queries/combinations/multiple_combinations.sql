    SELECT
        first_name,
        last_name
    FROM actor
    WHERE actor_id <= 10
EXCEPT
        SELECT
            first_name,
            last_name
        FROM actor
        WHERE
            actor_id >= 8 AND
            actor_id <= 12
    UNION
        SELECT
            first_name,
            last_name
        FROM actor
        WHERE actor_id >= 190
ORDER BY
    first_name,
    last_name;
