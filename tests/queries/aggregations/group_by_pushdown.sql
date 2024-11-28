SELECT *
FROM (
    SELECT
        length(f.title) AS title_length,
        count(*)
    FROM film f
    GROUP BY title_length
) AS s
WHERE
    title_length > 5 AND
    s.count > 3
ORDER BY
    s.count DESC,
    title_length;
