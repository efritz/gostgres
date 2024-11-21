SELECT
    film_id AS id,
    title AS movie_name,
    rental_rate AS price
FROM film
ORDER BY film_id
LIMIT 5;
