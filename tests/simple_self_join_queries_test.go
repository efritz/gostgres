package tests

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestSimpleSelfJoinQueries(t *testing.T) {
	testCases := []struct {
		query string
		want  autogold.Value
	}{
		{
			query: `
				SELECT * FROM employees e
				JOIN employees m ON m.employee_id = e.manager_id
			`,
			want: autogold.Want("TestSimpleSelfJoinQueries.join", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id, employee_id, first_name, last_name, email, manager_id, department_id)
    join
        alias as e
            access of employees
    with
        alias as m
            access of employees
    on m.employee_id = e.manager_id

Results:

 employee_id | first_name | last_name |              email             | manager_id | department_id | employee_id | first_name | last_name |            email           | manager_id | department_id
-------------+------------+-----------+--------------------------------+------------+---------------+-------------+------------+-----------+----------------------------+------------+---------------
           1 |   Annalisa |      Head |      annalisa.head@company.com |          1 |             1 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
           2 |    Clayton |  Mahaffey |   clayton.mahaffey@company.com |          4 |             2 |           4 |      Maria |    Warren |   maria.warren@company.com |          1 |             1
           3 |     Manuel |  Pattison |    manuel.pattison@company.com |          1 |             3 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
           4 |      Maria |    Warren |       maria.warren@company.com |          1 |             1 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
           5 |     Robert |    Medina |      robert.medina@company.com |          1 |             1 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
           6 |    Timothy |   Cornish |    timothy.cornish@company.com |          4 |             2 |           4 |      Maria |    Warren |   maria.warren@company.com |          1 |             1
           7 |      Linda |    Dollar |       linda.dollar@company.com |          1 |             1 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
           8 |  Frederick |  McLendon | frederick.mclendon@company.com |          4 |             2 |           4 |      Maria |    Warren |   maria.warren@company.com |          1 |             1
           9 |      Jimmy |  Barnette |     jimmy.barnette@company.com |          1 |             3 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
          10 |       Emma |    Howard |        emma.howard@company.com |          9 |             3 |           9 |      Jimmy |  Barnette | jimmy.barnette@company.com |          1 |             3
          11 |    Deborah |   Glasser |    deborah.glasser@company.com |          9 |             1 |           9 |      Jimmy |  Barnette | jimmy.barnette@company.com |          1 |             3
(11 rows)
`),
		},
		{
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
			want: autogold.Want("TestSimpleSelfJoinQueries.projection", `
Plan:

select (employee_id, last_name, m.employee_id as manager_id, m.first_name as manager_first_name, m.last_name as manager_last_name)
    join
        alias as e
            access of employees
    with
        alias as m
            access of employees
    on m.employee_id = e.manager_id

Results:

 employee_id | last_name | manager_id | manager_first_name | manager_last_name
-------------+-----------+------------+--------------------+-------------------
           1 |      Head |          1 |           Annalisa |              Head
           2 |  Mahaffey |          4 |              Maria |            Warren
           3 |  Pattison |          1 |           Annalisa |              Head
           4 |    Warren |          1 |           Annalisa |              Head
           5 |    Medina |          1 |           Annalisa |              Head
           6 |   Cornish |          4 |              Maria |            Warren
           7 |    Dollar |          1 |           Annalisa |              Head
           8 |  McLendon |          4 |              Maria |            Warren
           9 |  Barnette |          1 |           Annalisa |              Head
          10 |    Howard |          9 |              Jimmy |          Barnette
          11 |   Glasser |          9 |              Jimmy |          Barnette
(11 rows)
`),
		},
		{
			query: `
				SELECT * FROM employees e
				JOIN employees m ON m.employee_id = e.manager_id
				WHERE m.department_id != 3
			`,
			want: autogold.Want("TestSimpleSelfJoinQueries.filter", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id, employee_id, first_name, last_name, email, manager_id, department_id)
    join
        alias as e
            access of employees
    with
        alias as m
            access of employees
                filter: not employees.department_id = 3
    on m.employee_id = e.manager_id

Results:

 employee_id | first_name | last_name |              email             | manager_id | department_id | employee_id | first_name | last_name |           email           | manager_id | department_id
-------------+------------+-----------+--------------------------------+------------+---------------+-------------+------------+-----------+---------------------------+------------+---------------
           1 |   Annalisa |      Head |      annalisa.head@company.com |          1 |             1 |           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
           2 |    Clayton |  Mahaffey |   clayton.mahaffey@company.com |          4 |             2 |           4 |      Maria |    Warren |  maria.warren@company.com |          1 |             1
           3 |     Manuel |  Pattison |    manuel.pattison@company.com |          1 |             3 |           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
           4 |      Maria |    Warren |       maria.warren@company.com |          1 |             1 |           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
           5 |     Robert |    Medina |      robert.medina@company.com |          1 |             1 |           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
           6 |    Timothy |   Cornish |    timothy.cornish@company.com |          4 |             2 |           4 |      Maria |    Warren |  maria.warren@company.com |          1 |             1
           7 |      Linda |    Dollar |       linda.dollar@company.com |          1 |             1 |           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
           8 |  Frederick |  McLendon | frederick.mclendon@company.com |          4 |             2 |           4 |      Maria |    Warren |  maria.warren@company.com |          1 |             1
           9 |      Jimmy |  Barnette |     jimmy.barnette@company.com |          1 |             3 |           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
(9 rows)
`),
		},
		{
			query: `
				SELECT * FROM employees e
				JOIN employees m ON m.employee_id = e.manager_id
				ORDER BY m.last_name, e.last_name
			`,
			want: autogold.Want("TestSimpleSelfJoinQueries.order", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id, employee_id, first_name, last_name, email, manager_id, department_id)
    join
        alias as e
            access of employees
                order: employees.last_name
    with
        alias as m
            access of employees
                order: employees.last_name
    on m.employee_id = e.manager_id

Results:

 employee_id | first_name | last_name |              email             | manager_id | department_id | employee_id | first_name | last_name |            email           | manager_id | department_id
-------------+------------+-----------+--------------------------------+------------+---------------+-------------+------------+-----------+----------------------------+------------+---------------
           9 |      Jimmy |  Barnette |     jimmy.barnette@company.com |          1 |             3 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
           6 |    Timothy |   Cornish |    timothy.cornish@company.com |          4 |             2 |           4 |      Maria |    Warren |   maria.warren@company.com |          1 |             1
           7 |      Linda |    Dollar |       linda.dollar@company.com |          1 |             1 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
          11 |    Deborah |   Glasser |    deborah.glasser@company.com |          9 |             1 |           9 |      Jimmy |  Barnette | jimmy.barnette@company.com |          1 |             3
           1 |   Annalisa |      Head |      annalisa.head@company.com |          1 |             1 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
          10 |       Emma |    Howard |        emma.howard@company.com |          9 |             3 |           9 |      Jimmy |  Barnette | jimmy.barnette@company.com |          1 |             3
           2 |    Clayton |  Mahaffey |   clayton.mahaffey@company.com |          4 |             2 |           4 |      Maria |    Warren |   maria.warren@company.com |          1 |             1
           8 |  Frederick |  McLendon | frederick.mclendon@company.com |          4 |             2 |           4 |      Maria |    Warren |   maria.warren@company.com |          1 |             1
           5 |     Robert |    Medina |      robert.medina@company.com |          1 |             1 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
           3 |     Manuel |  Pattison |    manuel.pattison@company.com |          1 |             3 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
           4 |      Maria |    Warren |       maria.warren@company.com |          1 |             1 |           1 |   Annalisa |      Head |  annalisa.head@company.com |          1 |             1
(11 rows)
`),
		},
		{
			query: `
				SELECT * FROM employees e
				JOIN employees m ON m.employee_id = e.manager_id
				LIMIT 2
				OFFSET 3
			`,
			want: autogold.Want("TestSimpleSelfJoinQueries.limit", `
Plan:

select (employee_id, first_name, last_name, email, manager_id, department_id, employee_id, first_name, last_name, email, manager_id, department_id)
    limit 2
        offset 3
            join
                alias as e
                    access of employees
            with
                alias as m
                    access of employees
            on m.employee_id = e.manager_id

Results:

 employee_id | first_name | last_name |           email           | manager_id | department_id | employee_id | first_name | last_name |           email           | manager_id | department_id
-------------+------------+-----------+---------------------------+------------+---------------+-------------+------------+-----------+---------------------------+------------+---------------
           4 |      Maria |    Warren |  maria.warren@company.com |          1 |             1 |           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
           5 |     Robert |    Medina | robert.medina@company.com |          1 |             1 |           1 |   Annalisa |      Head | annalisa.head@company.com |          1 |             1
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
