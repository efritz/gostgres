-- SKIP
-- COUNT(*) doesn't work
-- GROUP BY length(title) has a type error

SELECT
    length(title) AS title_length,
    count(*) AS films_with_length
FROM film
GROUP BY length(title)
ORDER BY title_length;
