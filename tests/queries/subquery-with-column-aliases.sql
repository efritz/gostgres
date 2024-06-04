SELECT s.name || ' (' || s.lang || ')' AS normalized_title
FROM (
    SELECT f.title, l.name AS language_name
    FROM film f
    JOIN language l ON l.language_id = f.language_id
    ORDER BY f.title
    LIMIT 50
) s (name, lang);
