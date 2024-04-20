package tests

import (
	"testing"
)

func TestSimpleSelfJoinQueries(t *testing.T) {
	runTests(t, []TestCase{
		{
			name: "simple self-join",
			query: `
SELECT * 
FROM employees e
JOIN employees m ON m.employee_id = e.manager_id
`,
		},
		{
			name: "self-join with projection",
			query: `
SELECT
e.employee_id,
e.last_name,
m.employee_id AS manager_id,
m.first_name AS manager_first_name, 
m.last_name AS manager_last_name
FROM employees e
JOIN employees m ON m.employee_id = e.manager_id
`,
		},
		{
			name: "self-join with filter",
			query: `
SELECT * FROM employees e
JOIN employees m ON m.employee_id = e.manager_id
WHERE m.department_id != 3
`,
		},
		{
			name: "self-join with multi-colum order",
			query: `
SELECT * FROM employees e
JOIN employees m ON m.employee_id = e.manager_id
ORDER BY m.last_name, e.last_name
`,
		},
		{
			name: "self-join with limit and offset",
			query: `
SELECT * FROM employees e
JOIN employees m ON m.employee_id = e.manager_id
LIMIT 2 OFFSET 3
`,
		},
	})
}
