SELECT length(f.title) AS len, count(1)
FROM film f
GROUP BY len
ORDER BY len;
