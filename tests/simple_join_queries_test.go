package tests

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestSimpleJoinQueries(t *testing.T) {
	testCases := []struct {
		query string
		want  autogold.Value
	}{
		{
			query: `
				SELECT * FROM locations l
				JOIN regions r ON r.region_id = l.region_id
			`,
			want: autogold.Want("TestSimpleJoinQueries.join", `
Plan:

select (location_id, location_name, region_id, region_id, region_name)
    join using hash
        alias as l
            access of locations
    with
        alias as r
            access of regions
    on r.region_id = l.region_id

Results:

 location_id | location_name | region_id | region_id | region_name
-------------+---------------+-----------+-----------+-------------
           1 | San Francisco |         1 |         1 |          NA
           2 |       Toronto |         1 |         1 |          NA
           3 |      New York |         1 |         1 |          NA
           4 |     Barcelona |         2 |         2 |        EMEA
           5 |     Cape Town |         2 |         2 |        EMEA
           6 |     Guangzhou |         2 |         2 |        EMEA
(6 rows)
`),
		},
		{
			query: `
				SELECT
					location_name AS lname,
					r.region_name AS rname
				FROM locations l
				JOIN regions r ON r.region_id = l.region_id
			`,
			want: autogold.Want("TestSimpleJoinQueries.projection", `
Plan:

select (location_name as lname, r.region_name as rname)
    join using hash
        alias as l
            access of locations
    with
        alias as r
            access of regions
    on r.region_id = l.region_id

Results:

     lname     | rname
---------------+-------
 San Francisco |    NA
       Toronto |    NA
      New York |    NA
     Barcelona |  EMEA
     Cape Town |  EMEA
     Guangzhou |  EMEA
(6 rows)
`),
		},
		{
			query: `
				SELECT * FROM locations l
				JOIN regions r ON r.region_id = l.region_id AND region_name != 'NA'
			`,
			want: autogold.Want("TestSimpleJoinQueries.filter", `
Plan:

select (location_id, location_name, region_id, region_id, region_name)
    join using hash
        alias as l
            access of locations
    with
        alias as r
            access of regions
                filter: not region_name = NA
    on r.region_id = l.region_id

Results:

 location_id | location_name | region_id | region_id | region_name
-------------+---------------+-----------+-----------+-------------
           4 |     Barcelona |         2 |         2 |        EMEA
           5 |     Cape Town |         2 |         2 |        EMEA
           6 |     Guangzhou |         2 |         2 |        EMEA
(3 rows)
`),
		},
		{
			query: `
				SELECT * FROM locations l
				JOIN regions r ON r.region_id = l.region_id
				ORDER BY region_name, location_name DESC
			`,
			want: autogold.Want("TestSimpleJoinQueries.order", `
Plan:

select (location_id, location_name, region_id, region_id, region_name)
    order by region_name, location_name desc
        join using hash
            alias as l
                access of locations
                    order: location_name desc
        with
            alias as r
                access of regions
                    order: region_name
        on r.region_id = l.region_id

Results:

 location_id | location_name | region_id | region_id | region_name
-------------+---------------+-----------+-----------+-------------
           6 |     Guangzhou |         2 |         2 |        EMEA
           5 |     Cape Town |         2 |         2 |        EMEA
           4 |     Barcelona |         2 |         2 |        EMEA
           2 |       Toronto |         1 |         1 |          NA
           1 | San Francisco |         1 |         1 |          NA
           3 |      New York |         1 |         1 |          NA
(6 rows)
`),
		},
		{
			query: `
				SELECT * FROM locations
				JOIN regions
				LIMIT 2
				OFFSET 3
			`,
			want: autogold.Want("TestSimpleJoinQueries.limit", `
Plan:

select (location_id, location_name, region_id, region_id, region_name)
    limit 2
        offset 3
            join using nested loop
                access of locations
            with
                access of regions

Results:

 location_id | location_name | region_id | region_id | region_name
-------------+---------------+-----------+-----------+-------------
           2 |       Toronto |         1 |         2 |        EMEA
           3 |      New York |         1 |         1 |          NA
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
