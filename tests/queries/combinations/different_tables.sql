    SELECT
        first_name,
        last_name,
        'actor' AS source
    FROM actor
    WHERE actor_id < 5
UNION ALL
    SELECT
        first_name,
        last_name,
        'staff' AS source
    FROM staff
ORDER BY
    first_name,
    last_name,
    source;