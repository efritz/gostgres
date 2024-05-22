SELECT a.district, count(1)
FROM address a
GROUP BY a.district
ORDER BY a.count DESC
LIMIT 10;
