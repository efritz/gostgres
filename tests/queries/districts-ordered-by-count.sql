SELECT a.district, count(1)
FROM address a
GROUP BY a.district
ORDER BY count DESC, a.district
LIMIT 10;

-- NOTE: a.count should actually work too for some reason...?
