package tests

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestSimpleExceptQueries(t *testing.T) {
	testCases := []struct {
		query string
		want  autogold.Value
	}{
		{
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
			want: autogold.Want("TestSimpleExceptQueries.except.all", `
Plan:

select (location_id, location_name, region_id)
    order by region_id, location_id
        union
            select (location_id, location_name, region_id)
                access of locations
                    order: locations.region_id, locations.location_id
        with
            select (location_id, location_name, region_id)
                combination
                    select (location_id, location_name, region_id)
                        access of locations
                            order: locations.region_id, locations.location_id
                with
                    select (location_id, location_name, region_id)
                        join using hash
                            alias as l
                                access of locations
                        with
                            alias as r
                                access of regions
                                    filter: regions.region_name = NA
                        on r.region_id = l.region_id

Results:

 location_id | location_name | region_id
-------------+---------------+-----------
           1 | San Francisco |         1
           2 |       Toronto |         1
           3 |      New York |         1
           4 |     Barcelona |         2
           4 |     Barcelona |         2
           5 |     Cape Town |         2
           5 |     Cape Town |         2
           6 |     Guangzhou |         2
           6 |     Guangzhou |         2
(9 rows)
`),
		},
		{
			query: `
				SELECT
					employee_id AS id,
					e.last_name AS name
				FROM employees e
				WHERE department_id = 1
				EXCEPT (
					SELECT
						employee_id,
						last_name
					FROM employees
					WHERE manager_id != 1
				)
				ORDER BY id
			`,
			want: autogold.Want("TestSimpleExceptQueries.except.projection", `
Plan:

select (id, name)
    order by id
        combination
            select (employee_id as id, e.last_name as name)
                alias as e
                    access of employees
                        filter: department_id = 1
                        order: employee_id
        with
            select (employee_id, last_name)
                access of employees
                    filter: not manager_id = 1

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
					SELECT * FROM employees e
					WHERE department_id = 1
					EXCEPT (
						SELECT * FROM employees
						WHERE manager_id != 1
					)
				) s
				WHERE employee_id < 5
				ORDER by employee_id
			`,
			want: autogold.Want("TestSimpleExceptQueries.except.filter", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    order by employee_id
        alias as s
            select (employee_id, first_name, last_name, email, manager_id, department_id)
                combination
                    select (employee_id, first_name, last_name, email, manager_id, department_id)
                        alias as e
                            access of employees
                                filter: department_id = 1 and employees.employee_id < 5
                                order: employees.employee_id
                with
                    select (employee_id, first_name, last_name, email, manager_id, department_id)
                        access of employees
                            filter: not manager_id = 1

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
				SELECT * FROM employees e
				WHERE department_id = 1
				EXCEPT (
					SELECT * FROM employees
					WHERE manager_id != 1
				)
				ORDER by employee_id
				LIMIT 2 OFFSET 1
			`,
			want: autogold.Want("TestSimpleExceptQueries.except.limit", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    limit 2
        offset 1
            order by employee_id
                combination
                    select (employee_id, first_name, last_name, email, manager_id, department_id)
                        alias as e
                            access of employees
                                filter: department_id = 1
                                order: employees.employee_id
                with
                    select (employee_id, first_name, last_name, email, manager_id, department_id)
                        access of employees
                            filter: not manager_id = 1
                            order: employees.employee_id

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
			if err != nil {
				t.Fatalf("unexpected error running query: %s", err)
			}
			testCase.want.Equal(t, got)
		})
	}
}
