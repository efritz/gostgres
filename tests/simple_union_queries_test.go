package tests

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestSimpleUnionQueries(t *testing.T) {
	testCases := []struct {
		query string
		want  autogold.Value
	}{
		{
			query: `
				SELECT * FROM employees 
				WHERE department_id = 1
				UNION ALL (
					SELECT * FROM employees 
					WHERE manager_id = 1
				)
			`,
			want: autogold.Want("TestSimpleUnionQueries.union.all", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    union
        select (employee_id, first_name, last_name, email, manager_id, department_id)
            table scan of employees
                filter: department_id = 1
    with
        select (employee_id, first_name, last_name, email, manager_id, department_id)
            table scan of employees
                filter: manager_id = 1

Results:

 employee_id | first_name | last_name |            email            | manager_id | department_id
-------------+------------+-----------+-----------------------------+------------+---------------
           1 |   Annalisa |      Head |   annalisa.head@company.com |          1 |             1
           4 |      Maria |    Warren |    maria.warren@company.com |          1 |             1
           5 |     Robert |    Medina |   robert.medina@company.com |          1 |             1
           7 |      Linda |    Dollar |    linda.dollar@company.com |          1 |             1
          11 |    Deborah |   Glasser | deborah.glasser@company.com |          9 |             1
           1 |   Annalisa |      Head |   annalisa.head@company.com |          1 |             1
           3 |     Manuel |  Pattison | manuel.pattison@company.com |          1 |             3
           4 |      Maria |    Warren |    maria.warren@company.com |          1 |             1
           5 |     Robert |    Medina |   robert.medina@company.com |          1 |             1
           7 |      Linda |    Dollar |    linda.dollar@company.com |          1 |             1
           9 |      Jimmy |  Barnette |  jimmy.barnette@company.com |          1 |             3
(11 rows)
`),
		},
		{
			query: `
				SELECT
					employee_id AS id, 
					e.last_name AS name 
				FROM employees e 
				WHERE department_id = 1
				UNION (
					SELECT
						employee_id,
						last_name 
					FROM employees
					WHERE department_id = 2
				)
			`,
			want: autogold.Want("TestSimpleUnionQueries.union.projection", `
Plan:

select (id, name)
    union
        select (employee_id as id, e.last_name as name)
            alias as e
                table scan of employees
                    filter: department_id = 1
    with
        select (employee_id, last_name)
            table scan of employees
                filter: department_id = 2

Results:

 id |   name
----+----------
  1 |     Head
  4 |   Warren
  5 |   Medina
  7 |   Dollar
 11 |  Glasser
  2 | Mahaffey
  6 |  Cornish
  8 | McLendon
(8 rows)
`),
		},
		{
			query: `
				SELECT * FROM (
					SELECT * FROM employees 
					WHERE department_id = 1
					UNION (
						SELECT * FROM employees 
						WHERE department_id = 2
					)
				) s
				WHERE s.manager_id = 1
			`,
			want: autogold.Want("TestSimpleUnionQueries.union.filter", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    alias as s
        select (employee_id, first_name, last_name, email, manager_id, department_id)
            union
                select (employee_id, first_name, last_name, email, manager_id, department_id)
                    table scan of employees
                        filter: department_id = 1 and employees.manager_id = 1
            with
                select (employee_id, first_name, last_name, email, manager_id, department_id)
                    table scan of employees
                        filter: department_id = 2 and employees.manager_id = 1

Results:

 employee_id | first_name | last_name |           email           | manager_id | department_id
-------------+------------+-----------+---------------------------+------------+---------------
           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
           4 |      Maria |    Warren |  maria.warren@company.com |          1 |             1
           5 |     Robert |    Medina | robert.medina@company.com |          1 |             1
           7 |      Linda |    Dollar |  linda.dollar@company.com |          1 |             1
(4 rows)
`),
		},
		{
			query: `
				SELECT * FROM employees 
				WHERE department_id = 1
				UNION (
					SELECT * FROM employees 
					WHERE department_id = 2
				)
				ORDER BY email
			`,
			want: autogold.Want("TestSimpleUnionQueries.union.order", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    order by email
        union
            select (employee_id, first_name, last_name, email, manager_id, department_id)
                table scan of employees
                    filter: department_id = 1
        with
            select (employee_id, first_name, last_name, email, manager_id, department_id)
                table scan of employees
                    filter: department_id = 2

Results:

 employee_id | first_name | last_name |              email             | manager_id | department_id
-------------+------------+-----------+--------------------------------+------------+---------------
           1 |   Annalisa |      Head |      annalisa.head@company.com |          1 |             1
           2 |    Clayton |  Mahaffey |   clayton.mahaffey@company.com |          4 |             2
          11 |    Deborah |   Glasser |    deborah.glasser@company.com |          9 |             1
           8 |  Frederick |  McLendon | frederick.mclendon@company.com |          4 |             2
           7 |      Linda |    Dollar |       linda.dollar@company.com |          1 |             1
           4 |      Maria |    Warren |       maria.warren@company.com |          1 |             1
           5 |     Robert |    Medina |      robert.medina@company.com |          1 |             1
           6 |    Timothy |   Cornish |    timothy.cornish@company.com |          4 |             2
(8 rows)
`),
		},
		{
			query: `
				SELECT * FROM employees 
				WHERE department_id = 1
				UNION (
					SELECT * FROM employees 
					WHERE department_id = 2
				)
				LIMIT 2 OFFSET 3
			`,
			want: autogold.Want("TestSimpleUnionQueries.union.limit", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    limit 2
        offset 3
            union
                select (employee_id, first_name, last_name, email, manager_id, department_id)
                    table scan of employees
                        filter: department_id = 1
            with
                select (employee_id, first_name, last_name, email, manager_id, department_id)
                    table scan of employees
                        filter: department_id = 2

Results:

 employee_id | first_name | last_name |            email            | manager_id | department_id
-------------+------------+-----------+-----------------------------+------------+---------------
           7 |      Linda |    Dollar |    linda.dollar@company.com |          1 |             1
          11 |    Deborah |   Glasser | deborah.glasser@company.com |          9 |             1
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
