# TODO

## Schema features

- Support table constraints
- Support enums
- Support views
- Support materialized views
- Support ON CONFLICT
- Support additional DDL statements
- Support functions
- Support exclusion constraints
- Support ON DELETE/ON UPDATE of fkey constraints

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
- Non-hack planning for merge joins
- Implement Mark+Restore on indexes directly
- Break `Node` iface into optional components
- Combine alias and projection nodes (if possible)
