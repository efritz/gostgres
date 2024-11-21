SELECT min(replacement_cost), max(replacement_cost)
FROM film
-- Temporarily necessary
GROUP BY 1;
