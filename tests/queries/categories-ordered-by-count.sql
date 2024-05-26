SELECT c.name, count(1)
FROM film f
JOIN film_category fc ON fc.film_id = f.film_id
JOIN category c ON c.category_id = fc.category_id
GROUP BY c.name
-- "c.name" doesn't resolve
ORDER BY count, name;
