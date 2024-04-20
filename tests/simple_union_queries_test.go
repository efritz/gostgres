package tests

import (
	"testing"
)

func TestSimpleUnionQueries(t *testing.T) {
	runTests(t, []TestCase{
		{
			name: "simple union",
			query: `
SELECT * 
FROM employees 
WHERE department_id = 1
UNION ALL (
	SELECT * 
	FROM employees 
	WHERE manager_id = 1
)
			`,
		},
		{
			name: "union with projection",
			query: `
SELECT employee_id AS id, e.last_name AS name 
FROM employees e 
WHERE department_id = 1
UNION (
	SELECT employee_id, last_name 
	FROM employees
	WHERE department_id = 2
)
			`,
		},
		{
			name: "union with filter",
			query: `
SELECT * FROM (
	SELECT * 
	FROM employees 
	WHERE department_id = 1
	UNION (
		SELECT * 
		FROM employees 
		WHERE department_id = 2
	)
) s
WHERE s.manager_id = 1
			`,
		},
		{
			name: "union with order",
			query: `
SELECT * 
FROM employees 
WHERE department_id = 1
UNION (
	SELECT * 
	FROM employees 
	WHERE department_id = 2
)
ORDER BY email
			`,
		},
		{
			name: "union with limit and offset",
			query: `
SELECT * 
FROM employees 
WHERE department_id = 1
UNION (
	SELECT * 
	FROM employees 
	WHERE department_id = 2
)
LIMIT 2 OFFSET 3
			`,
		},
	})
}
