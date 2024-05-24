SELECT a.district, count(1)
FROM address a
GROUP BY a.district
ORDER BY a.count DESC, a.district
LIMIT 10;
