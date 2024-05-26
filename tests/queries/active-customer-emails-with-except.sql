SELECT s.* FROM (
    SELECT email FROM customer
    EXCEPT
    SELECT email FROM customer WHERE active = 0
) s
ORDER BY s.email;
