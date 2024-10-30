SELECT *
FROM film_category fc
JOIN (SELECT category_id, name FROM category c) cat(id, name) ON cat.id = fc.category_id;




CREATE TABLE ys (id serial primary key, x integer);
CREATE TABLE zs (id serial primary key, q integer, w text);
SELECT s.x FROM (SELECT * FROM ys y JOIN zs z ON z.y_id = y.id) s;
