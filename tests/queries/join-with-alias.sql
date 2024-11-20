SELECT * 
FROM film_category fc
JOIN (SELECT category_id, name FROM category c) cat(id, name) ON cat.id = fc.category_id
ORDER BY fc.film_id
LIMIT 50;
