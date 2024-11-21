SELECT
    f.title AS film_title,
    f.rental_rate + f.replacement_cost AS total_cost
FROM film f
ORDER BY total_cost DESC, f.title;
