SELECT f.title, a.first_name || ' ' || a.last_name
FROM film f
JOIN film_actor fa ON fa.film_id = f.film_id
JOIN actor a ON a.actor_id = fa.actor_id
ORDER BY a.last_name DESC, a.first_name DESC
LIMIT 50;
