package tests

import (
	"testing"
)

func TestSimpleJoinQueries(t *testing.T) {
	runTests(t, []TestCase{
		{
			name: "simple join",
			query: `
SELECT * FROM locations l
JOIN regions r ON r.region_id = l.region_id
`,
		},
		{
			name: "simple join with projection",
			query: `
SELECT location_name AS lname, r.region_name AS rname
FROM locations l
JOIN regions r ON r.region_id = l.region_id
`,
		},
		{
			name: "simple join with multiple conditions",
			query: `
SELECT *
FROM locations l
JOIN regions r ON r.region_id = l.region_id AND region_name != 'NA'
`,
		},
		{
			name: "simple join with multi-column order",
			query: `
SELECT *
FROM locations l
JOIN regions r ON r.region_id = l.region_id
ORDER BY region_name, location_name DESC
`,
		},
		{
			name: "implicit cross join with limit and offset",
			query: `
SELECT *
FROM locations
JOIN regions
LIMIT 2
OFFSET 3
`,
		},
	})
}
