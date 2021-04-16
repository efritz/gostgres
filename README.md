# Gostgres

Postgres in Go over a short vacation.

## TODO

- Short-term
    - Add types to values and rows
    - Support CTEs
    - Support builtin functions
    - Support DISTINCT (ON)
    - Support GROUP BY and HAVING
    - Support UNION/INTERSECT/EXCEPT
    - Support multiple sort expressions
    - Support sort directions
    - Support inserts with explicit column refs
    - Support UPDATE
    - Support DELETE
    - Support RETURNING

- Medium-term
    - Support default values
    - Support check constraints
    - Support indexes
    - Support inserts on conflict
    - Fetch optimization
    - Outer joins
    - Non-nested-loop joins

- Long-term
    - Write to disk
    - Recursive CTEs
    - Window queries
    - DDL
    - Networking
    - Multiple clients
    - Transactions
    - Triggers
    - WAL
