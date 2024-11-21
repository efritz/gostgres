-- SKIP
-- COUNT(*) doesn't work

SELECT
    length(title) AS title_length,
    count(*) AS films_with_length
FROM film
GROUP BY title_length
ORDER BY title_length;
