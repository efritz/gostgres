package tests

import (
	"testing"
)

func TestSimpleSelectQueries(t *testing.T) {
	runTests(t, []TestCase{
		{
			name: "select wildcard",
			query: `
SELECT *
FROM employees
`,
		},
		{
			name: "select with projection",
			query: `
SELECT first_name AS fname, e.last_name AS lname, email
FROM employees e
`,
		},
		{
			name: "select with simple filter",
			query: `
SELECT *
FROM employees
WHERE department_id = 3
`,
		},
		{
			name: "select with simple order",
			query: `
SELECT *
FROM employees
ORDER BY employee_id
`,
		},
		{
			name: "select with multi-column order",
			query: `
SELECT *
FROM employees
ORDER BY manager_id, email DESC
`,
		},
		{
			name: "select with limit and offset (ordered for stability)",
			query: `
SELECT *
FROM employees
ORDER BY email
LIMIT 2 OFFSET 3
`,
		},
	})
}
