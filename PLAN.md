- Add internal/unique id to all fields/rows
- Add static mapping from table.col -> field id

- Split named expression and resolved named expression
- Ensure resolution context tracks resolved entries in symtab
- Rewrite all query expressions to use resolved named expression
- Rewrite field lookup in tuples to use ids

- Add resolution step for DDL statements

- Get rid of alias node
- Move most of projection into resolution step
