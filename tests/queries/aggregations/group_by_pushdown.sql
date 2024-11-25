SELECT *
FROM (
    SELECT
        length(f.title) AS title_length,
        count(*) AS films_with_length
    FROM film f
    GROUP BY title_length
    ORDER BY title_length
) AS s
WHERE s.count > 3
ORDER BY s.count DESC;
