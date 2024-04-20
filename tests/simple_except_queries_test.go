package tests

import (
	"testing"
)

func TestSimpleExceptQueries(t *testing.T) {
	runTests(t, []TestCase{
		{
			name: "union all and except all",
			query: `
SELECT * FROM locations
UNION ALL
SELECT * FROM locations
EXCEPT ALL
(
	SELECT l.* FROM locations l
	JOIN regions r ON r.region_id = l.region_id
	WHERE r.region_name = 'NA'
)
ORDER BY region_id, location_id
`,
		},
		{
			name: "except with projection and order",
			query: `
SELECT employee_id AS id, e.last_name AS name
FROM employees e
WHERE department_id = 1
EXCEPT (
	SELECT employee_id, last_name
	FROM employees
	WHERE manager_id != 1
)
ORDER BY id
`,
		},
		{
			name: "except with filter and order",
			query: `
SELECT * 
FROM (
	SELECT * 
	FROM employees e
	WHERE department_id = 1
	EXCEPT (
		SELECT * 
		FROM employees
		WHERE manager_id != 1
	)
) s
WHERE employee_id < 5
ORDER by employee_id
`,
		},
		{
			name: "except with limit and offset",
			query: `
SELECT *
FROM employees e
WHERE department_id = 1
EXCEPT (
	SELECT *
	FROM employees
	WHERE manager_id != 1
)
ORDER by employee_id
LIMIT 2 OFFSET 1
`,
		},
	})
}
