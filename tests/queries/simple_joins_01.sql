SELECT f.title, f.rating, c.name
FROM film f
JOIN film_category fc ON fc.film_id = f.film_id
JOIN category c ON c.category_id = fc.category_id
WHERE rating = 'R'
ORDER BY f.title
LIMIT 20;
