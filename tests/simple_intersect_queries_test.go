package tests

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleIntersectQueries(t *testing.T) {
	testCases := []struct {
		query string
		want  autogold.Value
	}{
		{
			query: `
				SELECT * FROM locations
				INTERSECT ALL
				(
					SELECT l.* FROM locations l
					JOIN regions r ON r.region_id = l.region_id
					WHERE r.region_name = 'NA'
				)
				ORDER BY region_id, location_id
			`,
			want: autogold.Want("TestSimpleIntersectQueries.intersect.all", `
Plan:

select (location_id, location_name, region_id)
    order by region_id, location_id
        combination
            select (location_id, location_name, region_id)
                table scan of locations
        with
            select (location_id, location_name, region_id)
                join using nested loop
                    alias as l
                        table scan of locations
                with
                    alias as r
                        table scan of regions
                            filter: regions.region_name = NA and regions.region_id = l.region_id
                on r.region_id = l.region_id

Results:

 location_id | location_name | region_id
-------------+---------------+-----------
           1 | San Francisco |         1
           1 | San Francisco |         1
           2 |       Toronto |         1
           2 |       Toronto |         1
           3 |      New York |         1
           3 |      New York |         1
(6 rows)
`),
		},
		{
			query: `
				SELECT
					employee_id AS id,
					e.last_name AS name
				FROM employees e
				WHERE department_id = 1
				INTERSECT (
					SELECT
						employee_id,
						last_name
					FROM employees
					WHERE manager_id = 1
				)
				ORDER BY id
			`,
			want: autogold.Want("TestSimpleIntersectQueries.intersect.projection", `
Plan:

select (id, name)
    order by id
        combination
            select (employee_id as id, e.last_name as name)
                alias as e
                    table scan of employees
                        filter: department_id = 1
        with
            select (employee_id, last_name)
                table scan of employees
                    filter: manager_id = 1

Results:

 id |  name
----+--------
  1 |   Head
  4 | Warren
  5 | Medina
  7 | Dollar
(4 rows)
`),
		},
		{
			query: `
				SELECT * FROM (
					SELECT * FROM employees 
					WHERE department_id = 1
					INTERSECT (
						SELECT * FROM employees 
						WHERE manager_id = 1
					)
				) s
				WHERE employee_id < 5
				ORDER BY employee_id
			`,
			want: autogold.Want("TestSimpleIntersectQueries.intersect.filter", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    order by employee_id
        alias as s
            select (employee_id, first_name, last_name, email, manager_id, department_id)
                combination
                    select (employee_id, first_name, last_name, email, manager_id, department_id)
                        table scan of employees
                            filter: department_id = 1 and employees.employee_id < 5
                with
                    select (employee_id, first_name, last_name, email, manager_id, department_id)
                        table scan of employees
                            filter: manager_id = 1 and employees.employee_id < 5

Results:

 employee_id | first_name | last_name |           email           | manager_id | department_id
-------------+------------+-----------+---------------------------+------------+---------------
           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
           4 |      Maria |    Warren |  maria.warren@company.com |          1 |             1
(2 rows)
`),
		},
		{
			query: `
				SELECT * FROM employees 
				WHERE department_id = 1
				INTERSECT (
					SELECT * FROM employees 
					WHERE manager_id = 1
				)
				ORDER BY employee_id
				LIMIT 2 OFFSET 1
			`,
			want: autogold.Want("TestSimpleIntersectQueries.intersect.limit", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    limit 2
        offset 1
            order by employee_id
                combination
                    select (employee_id, first_name, last_name, email, manager_id, department_id)
                        table scan of employees
                            filter: department_id = 1
                with
                    select (employee_id, first_name, last_name, email, manager_id, department_id)
                        table scan of employees
                            filter: manager_id = 1

Results:

 employee_id | first_name | last_name |           email           | manager_id | department_id
-------------+------------+-----------+---------------------------+------------+---------------
           4 |      Maria |    Warren |  maria.warren@company.com |          1 |             1
           5 |     Robert |    Medina | robert.medina@company.com |          1 |             1
(2 rows)
`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.want.Name(), func(t *testing.T) {
			got, err := testQuery(testCase.query)
			require.NoError(t, err)
			assert.Equal(t, testCase.want, got)
		})
	}
}
