SELECT
    rating,
    rental_rate,
    count(*) AS film_count
FROM film
GROUP BY rating, rental_rate
ORDER BY rating, rental_rate;
