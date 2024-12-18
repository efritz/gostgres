SELECT v FROM (VALUES (1), (2), (2), (3), (3), (3), (4), (4), (4), (4)) AS t(v)
UNION ALL
SELECT * FROM (VALUES (1), (3), (3), (5))
ORDER BY v;
