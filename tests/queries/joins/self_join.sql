SELECT
    a1.actor_id,
    a1.first_name || ' ' || a1.last_name AS actor_name,
    a2.actor_id AS similar_actor_id,
    a2.first_name || ' ' || a2.last_name AS similar_actor_name,
    a1.last_name AS shared_last_name
FROM actor a1
JOIN actor a2 ON
    a1.last_name = a2.last_name AND
    a1.actor_id < a2.actor_id
ORDER BY
    a1.last_name,
    a1.actor_id
LIMIT 5;
