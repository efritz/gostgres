package tests

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestSimpleSelectQueries(t *testing.T) {
	testCases := []struct {
		query string
		want  autogold.Value
	}{
		{
			query: `
				SELECT * FROM employees
			`,
			want: autogold.Want("TestSimpleSelectQueries.select", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    table scan of employees

Results:

 employee_id | first_name | last_name |              email             | manager_id | department_id
-------------+------------+-----------+--------------------------------+------------+---------------
           1 |   Annalisa |      Head |      annalisa.head@company.com |          1 |             1
           2 |    Clayton |  Mahaffey |   clayton.mahaffey@company.com |          4 |             2
           3 |     Manuel |  Pattison |    manuel.pattison@company.com |          1 |             3
           4 |      Maria |    Warren |       maria.warren@company.com |          1 |             1
           5 |     Robert |    Medina |      robert.medina@company.com |          1 |             1
           6 |    Timothy |   Cornish |    timothy.cornish@company.com |          4 |             2
           7 |      Linda |    Dollar |       linda.dollar@company.com |          1 |             1
           8 |  Frederick |  McLendon | frederick.mclendon@company.com |          4 |             2
           9 |      Jimmy |  Barnette |     jimmy.barnette@company.com |          1 |             3
          10 |       Emma |    Howard |        emma.howard@company.com |          9 |             3
          11 |    Deborah |   Glasser |    deborah.glasser@company.com |          9 |             1
(11 rows)
`),
		},
		{
			query: `
				SELECT
					first_name AS fname,
					e.last_name AS lname,
					email
				FROM employees e
			`,
			want: autogold.Want("TestSimpleSelectQueries.projection", `
Plan:

select (first_name as fname, e.last_name as lname, email)
    alias as e
        table scan of employees

Results:

   fname   |   lname  |              email
-----------+----------+--------------------------------
  Annalisa |     Head |      annalisa.head@company.com
   Clayton | Mahaffey |   clayton.mahaffey@company.com
    Manuel | Pattison |    manuel.pattison@company.com
     Maria |   Warren |       maria.warren@company.com
    Robert |   Medina |      robert.medina@company.com
   Timothy |  Cornish |    timothy.cornish@company.com
     Linda |   Dollar |       linda.dollar@company.com
 Frederick | McLendon | frederick.mclendon@company.com
     Jimmy | Barnette |     jimmy.barnette@company.com
      Emma |   Howard |        emma.howard@company.com
   Deborah |  Glasser |    deborah.glasser@company.com
(11 rows)
`),
		},
		{
			query: `
				SELECT * FROM employees
				WHERE department_id = 3
			`,
			want: autogold.Want("TestSimpleSelectQueries.filter", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    table scan of employees
        filter: department_id = 3

Results:

 employee_id | first_name | last_name |            email            | manager_id | department_id
-------------+------------+-----------+-----------------------------+------------+---------------
           3 |     Manuel |  Pattison | manuel.pattison@company.com |          1 |             3
           9 |      Jimmy |  Barnette |  jimmy.barnette@company.com |          1 |             3
          10 |       Emma |    Howard |     emma.howard@company.com |          9 |             3
(3 rows)
`),
		},
		{
			query: `
				SELECT * FROM employees
				ORDER BY manager_id, email DESC
			`,
			want: autogold.Want("TestSimpleSelectQueries.order", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    order by manager_id, email desc
        table scan of employees

Results:

 employee_id | first_name | last_name |              email             | manager_id | department_id
-------------+------------+-----------+--------------------------------+------------+---------------
           5 |     Robert |    Medina |      robert.medina@company.com |          1 |             1
           4 |      Maria |    Warren |       maria.warren@company.com |          1 |             1
           3 |     Manuel |  Pattison |    manuel.pattison@company.com |          1 |             3
           7 |      Linda |    Dollar |       linda.dollar@company.com |          1 |             1
           9 |      Jimmy |  Barnette |     jimmy.barnette@company.com |          1 |             3
           1 |   Annalisa |      Head |      annalisa.head@company.com |          1 |             1
           6 |    Timothy |   Cornish |    timothy.cornish@company.com |          4 |             2
           8 |  Frederick |  McLendon | frederick.mclendon@company.com |          4 |             2
           2 |    Clayton |  Mahaffey |   clayton.mahaffey@company.com |          4 |             2
          10 |       Emma |    Howard |        emma.howard@company.com |          9 |             3
          11 |    Deborah |   Glasser |    deborah.glasser@company.com |          9 |             1
(11 rows)
`),
		},
		{
			query: `
				SELECT * FROM employees
				ORDER BY email
				LIMIT 2
				OFFSET 3
			`,
			want: autogold.Want("TestSimpleSelectQueries.limit", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id)
    limit 2
        offset 3
            order by email
                table scan of employees

Results:

 employee_id | first_name | last_name |              email             | manager_id | department_id
-------------+------------+-----------+--------------------------------+------------+---------------
          10 |       Emma |    Howard |        emma.howard@company.com |          9 |             3
           8 |  Frederick |  McLendon | frederick.mclendon@company.com |          4 |             2
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
