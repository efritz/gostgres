package repl

import (
	"github.com/efritz/gostgres/internal/relations"
	"github.com/efritz/gostgres/internal/shared"
)

var tables = map[string]*relations.Table{
	"employees":   employeesTable,
	"departments": departmentsTable,
	"locations":   locationsTable,
	"regions":     regionsTable,
}

var employeesTable = relations.NewTable(shared.Rows{
	Fields: []shared.Field{
		{RelationName: "employees", Name: "employee_id"},
		{RelationName: "employees", Name: "first_name"},
		{RelationName: "employees", Name: "last_name"},
		{RelationName: "employees", Name: "email"},
		{RelationName: "employees", Name: "manager_id"},
		{RelationName: "employees", Name: "department_id"},
	},
	Values: [][]interface{}{
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
})

var departmentsTable = relations.NewTable(shared.Rows{
	Fields: []shared.Field{
		{RelationName: "departments", Name: "department_id"},
		{RelationName: "departments", Name: "department_name"},
		{RelationName: "departments", Name: "location_id"},
	},
	Values: [][]interface{}{
		{1, "Team A", 1},
		{2, "Team B", 1},
		{3, "Team C", 4},
	},
})

var locationsTable = relations.NewTable(shared.Rows{
	Fields: []shared.Field{
		{RelationName: "locations", Name: "location_id"},
		{RelationName: "locations", Name: "location_name"},
		{RelationName: "locations", Name: "region_id"},
	},
	Values: [][]interface{}{
		{1, "San Francisco", 1},
		{2, "Toronto", 1},
		{3, "New York", 1},
		{4, "Barcelona", 2},
		{5, "Cape Town", 2},
		{6, "Guangzhou", 2},
	},
})

var regionsTable = relations.NewTable(shared.Rows{
	Fields: []shared.Field{
		{RelationName: "regions", Name: "region_id"},
		{RelationName: "regions", Name: "region_name"},
	},
	Values: [][]interface{}{
		{1, "NA"},
		{2, "EMEA"},
	},
})
