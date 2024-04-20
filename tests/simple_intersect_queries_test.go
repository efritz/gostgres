package tests

import (
	"testing"
)

func TestSimpleIntersectQueries(t *testing.T) {
	runTests(t, []TestCase{
		{
			name: "simple intersect",
			query: `
SELECT * 
FROM locations
INTERSECT ALL
(
	SELECT l.* 
	FROM locations l
	JOIN regions r ON r.region_id = l.region_id
	WHERE r.region_name = 'NA'
)
ORDER BY region_id, location_id
`,
		},
		{
			name: "intersect with projection and order",
			query: `
SELECT employee_id AS id, e.last_name AS name
FROM employees e
WHERE department_id = 1
INTERSECT (
	SELECT employee_id, last_name
	FROM employees
	WHERE manager_id = 1
)
ORDER BY id
`,
		},
		{
			name: "intersect with filter and order",
			query: `
SELECT * 
FROM (
	SELECT * 
	FROM employees 
	WHERE department_id = 1
	INTERSECT (
		SELECT * 
		FROM employees 
		WHERE manager_id = 1
	)
) s
WHERE employee_id < 5
ORDER BY employee_id
`,
		},
		{
			name: "intersect with limit and offset",
			query: `
SELECT * 
FROM employees 
WHERE department_id = 1
INTERSECT (
	SELECT * 
	FROM employees 
	WHERE manager_id = 1
)
ORDER BY employee_id
LIMIT 2 OFFSET 1
`,
		},
	})
}
