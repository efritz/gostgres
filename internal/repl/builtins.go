package repl

import (
	nodes "github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/shared"
)

var tables map[string]*nodes.Table

func init() {
	employeesRows, err := shared.NewRowsWithValues(
		[]shared.Field{
			shared.NewField("employees", "employee_id", shared.TypeKindNumeric, false),
			shared.NewField("employees", "first_name", shared.TypeKindText, false),
			shared.NewField("employees", "last_name", shared.TypeKindText, false),
			shared.NewField("employees", "email", shared.TypeKindText, false),
			shared.NewField("employees", "manager_id", shared.TypeKindNumeric, false),
			shared.NewField("employees", "department_id", shared.TypeKindNumeric, false),
		},
		[][]interface{}{
			{1, "Annalisa", "Head", "annalisa.head@company.com", 1, 1},
			{2, "Clayton", "Mahaffey", "clayton.mahaffey@company.com", 4, 2},
			{3, "Manuel", "Pattison", "manuel.pattison@company.com", 1, 3},
			{4, "Maria", "Warren", "maria.warren@company.com", 1, 1},
			{5, "Robert", "Medina", "robert.medina@company.com", 1, 1},
			{6, "Timothy", "Cornish", "timothy.cornish@company.com", 4, 2},
			{7, "Linda", "Dollar", "linda.dollar@company.com", 1, 1},
			{8, "Frederick", "McLendon", "frederick.mclendon@company.com", 4, 2},
			{9, "Jimmy", "Barnette", "jimmy.barnette@company.com", 1, 3},
			{10, "Emma", "Howard", "emma.howard@company.com", 9, 3},
			{11, "Deborah", "Glasser", "deborah.glasser@company.com", 9, 1},
		},
	)
	if err != nil {
		panic(err.Error())
	}
	employeesTable, err := nodes.NewTable("employees", employeesRows)
	if err != nil {
		panic(err.Error())
	}

	departmentsRows, err := shared.NewRowsWithValues(
		[]shared.Field{
			shared.NewField("departments", "department_id", shared.TypeKindNumeric, false),
			shared.NewField("departments", "department_name", shared.TypeKindText, false),
			shared.NewField("departments", "location_id", shared.TypeKindNumeric, false),
		},
		[][]interface{}{
			{1, "Team A", 1},
			{2, "Team B", 1},
			{3, "Team C", 4},
		},
	)
	if err != nil {
		panic(err.Error())
	}
	departmentsTable, err := nodes.NewTable("departments", departmentsRows)
	if err != nil {
		panic(err.Error())
	}

	locationsRows, err := shared.NewRowsWithValues(
		[]shared.Field{
			shared.NewField("locations", "location_id", shared.TypeKindNumeric, false),
			shared.NewField("locations", "location_name", shared.TypeKindText, false),
			shared.NewField("locations", "region_id", shared.TypeKindNumeric, false),
		},
		[][]interface{}{
			{1, "San Francisco", 1},
			{2, "Toronto", 1},
			{3, "New York", 1},
			{4, "Barcelona", 2},
			{5, "Cape Town", 2},
			{6, "Guangzhou", 2},
		},
	)
	if err != nil {
		panic(err.Error())
	}
	locationsTable, err := nodes.NewTable("locations", locationsRows)
	if err != nil {
		panic(err.Error())
	}

	regionsRows, err := shared.NewRowsWithValues(
		[]shared.Field{
			shared.NewField("regions", "region_id", shared.TypeKindNumeric, false),
			shared.NewField("regions", "region_name", shared.TypeKindText, false),
		},
		[][]interface{}{
			{1, "NA"},
			{2, "EMEA"},
		},
	)
	if err != nil {
		panic(err.Error())
	}
	regionsTable, err := nodes.NewTable("regions", regionsRows)
	if err != nil {
		panic(err.Error())
	}

	tables = map[string]*nodes.Table{
		"employees":   employeesTable,
		"departments": departmentsTable,
		"locations":   locationsTable,
		"regions":     regionsTable,
	}
}
