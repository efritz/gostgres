# TODO

## Schema features

- Support foreign keys
- Support enums
- Support views
- Support materialized views
- Support ON CONFLICT
- Support additional DDL statements
- Support functions
- Support exclusion constraints

## Query features

- Support Non-inner joins
- Support CTEs
- Support recursive CTEs
- Support DISTINCT (ON)
- Support GROUP BY and HAVING
- Support window queries
- Support subquery expressions (EXISTS/IN/NOT IN/ANY/SOME/ALL)
- Support row comparisons (IN/NOT IN/ANY/SOME/ALL)
- Support TRUNCATE

## Internal features

- Disk persistence
- WAL logging
- Networking
- Multiple clients
- Transactions
- Triggers

## Tech debt

- Rewrite non-null types into constraints
- Move default out of fields and into table column definition
- Non-hack planning for merge joins
- Implement Mark+Restore on indexes directly
- Break `Node` iface into optional components
- Combine alias and projection nodes (if possible)