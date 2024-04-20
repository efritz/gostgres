# Gostgres

Postgres in Go as a learning exercise.

## Try me out

Simply run `go build ./cmd/repl && ./repl` to drop into a psql-like shell where you can issue SQL commands to an in-memory database.

Currently this shell pre-loads a number of relations and data (`employees`, `departments`, `locations`, and `regions`).

```
$ go build ./cmd/repl && ./repl
gostgres ❯ select * from employees where department_id = 2 order by last_name;
 employee_id | first_name | last_name |              email             | manager_id | department_id
-------------+------------+-----------+--------------------------------+------------+---------------
           6 |    Timothy |   Cornish |    timothy.cornish@company.com |          4 |             2
           2 |    Clayton |  Mahaffey |   clayton.mahaffey@company.com |          4 |             2
           8 |  Frederick |  McLendon | frederick.mclendon@company.com |          4 |             2
(3 rows)

gostgres ❯ select l.*, r.region_name from locations l join regions r on l.region_id = r.region_id;
 location_id | location_name | region_id | region_name
-------------+---------------+-----------+-------------
           1 | San Francisco |         1 |          NA
           2 |       Toronto |         1 |          NA
           3 |      New York |         1 |          NA
           4 |     Barcelona |         2 |        EMEA
           5 |     Cape Town |         2 |        EMEA
           6 |     Guangzhou |         2 |        EMEA
(6 rows)

Time: 106.244µs
```

Currently, Gostgres supports `SELECT`, `INSERT`, `UPDATE`, and `DELETE` to varying degrees.

## TODO

- Short-term
    - Support CTEs
    - Support builtin functions
    - Support DISTINCT (ON)
    - Support GROUP BY and HAVING
    - Support TRUNCATE
    - Support subquery expressions (EXISTS/IN/NOT IN/ANY/SOME/ALL)
    - Support row comparisons (IN/NOT IN/ANY/SOME/ALL)

- Medium-term
    - Primary keys
    - Foreign keys
    - Nullability
    - Enums
    - Constraints
    - Views
    - Materialized views
    - Default values
    - ON CONFLICT
    - Non-inner joins

- Long-term
    - Recursive CTEs
    - Window queries
    - Write to disk
    - DDL
    - Networking
    - Multiple clients
    - Functions
    - Transactions
    - Triggers
    - WAL

- Tech debt
    - Non-hack planning for merge joins
    - Implement Mark+Restore on indexes directly
    - Break `Node` iface into optional components
    - Combine alias and projection nodes (if possible)
