package tests

import (
	"fmt"
	"path/filepath"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/loader"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

func CreateStandardTestTables(root string) (map[string]*table.Table, error) {
	loaders := map[string]func(string) (*table.Table, error){
		"employees":   createEmployeesTable,
		"departments": createDepartmentsTable,
		"locations":   createLocationsTable,
		"regions":     createRegionsTable,
		"k1":          createK1Table,
		"k2":          createK2Table,
	}

	tables := make(map[string]*table.Table)
	for name, loader := range loaders {
		table, err := loader(root)
		if err != nil {
			return nil, err
		}

		tables[name] = table
	}

	return tables, nil
}

func createEmployeesTable(root string) (*table.Table, error) {
	employeeID := shared.NewField("employees", "employee_id", shared.TypeKindNumeric)
	firstName := shared.NewField("employees", "first_name", shared.TypeKindText)
	last_name := shared.NewField("employees", "last_name", shared.TypeKindText)
	email := shared.NewField("employees", "email", shared.TypeKindText)
	managerID := shared.NewField("employees", "manager_id", shared.TypeKindNumeric)
	departmentID := shared.NewField("employees", "department_id", shared.TypeKindNumeric)

	table, err := loader.NewTableFromCSV("employees", csvFilepath(root, "employees"), []shared.Field{
		employeeID,
		firstName,
		last_name,
		email,
		managerID,
		departmentID,
	})
	if err != nil {
		return nil, err
	}

	// btree index on (last_name, first_name, employee_id)
	if err := table.AddIndex(indexes.NewBTreeIndex(
		"employees_last_name_first_name_employee_id",
		table.Name(),
		[]expressions.ExpressionWithDirection{
			{Expression: expressions.NewNamed(last_name)},
			{Expression: expressions.NewNamed(firstName)},
			{Expression: expressions.NewNamed(employeeID)},
		},
	)); err != nil {
		return nil, err
	}

	// hash index on first name
	if err := table.AddIndex(indexes.NewHashIndex(
		"employees_first_name",
		table.Name(),
		expressions.NewNamed(firstName),
	)); err != nil {
		return nil, err
	}

	// hash index last_name, partial where manager_id <= 4
	if err := table.AddIndex(indexes.NewPartialIndex(
		indexes.NewHashIndex(
			"employees_last_name_manager_id",
			table.Name(),
			expressions.NewNamed(last_name),
		),
		expressions.NewLessThanEquals(expressions.NewNamed(managerID), expressions.NewConstant(4)),
	)); err != nil {
		return nil, err
	}

	return table, nil
}

func createDepartmentsTable(root string) (*table.Table, error) {
	departmentID := shared.NewField("departments", "department_id", shared.TypeKindNumeric)
	departmentName := shared.NewField("departments", "department_name", shared.TypeKindText)
	locationID := shared.NewField("departments", "location_id", shared.TypeKindNumeric)

	table, err := loader.NewTableFromCSV("departments", csvFilepath(root, "departments"), []shared.Field{
		departmentID,
		departmentName,
		locationID,
	})
	if err != nil {
		return nil, err
	}

	// hash index on department_id
	if err := table.AddIndex(indexes.NewHashIndex(
		"departments_department_id",
		table.Name(),
		expressions.NewNamed(departmentID),
	)); err != nil {
		return nil, err
	}

	return table, nil
}

func createLocationsTable(root string) (*table.Table, error) {
	locationID := shared.NewField("locations", "location_id", shared.TypeKindNumeric)
	locationName := shared.NewField("locations", "location_name", shared.TypeKindText)
	regionID := shared.NewField("locations", "region_id", shared.TypeKindNumeric)

	return loader.NewTableFromCSV("locations", csvFilepath(root, "locations"), []shared.Field{
		locationID,
		locationName,
		regionID,
	})
}

func createRegionsTable(root string) (*table.Table, error) {
	regionID := shared.NewField("regions", "region_id", shared.TypeKindNumeric)
	regionName := shared.NewField("regions", "region_name", shared.TypeKindText)

	return loader.NewTableFromCSV("regions", csvFilepath(root, "regions"), []shared.Field{
		regionID,
		regionName,
	})
}

func createK1Table(root string) (*table.Table, error) {
	name := shared.NewField("k1", "name", shared.TypeKindText)
	id := shared.NewField("k1", "id", shared.TypeKindNumeric)

	table, err := loader.NewTableFromCSV("k1", csvFilepath(root, "k1"), []shared.Field{
		name,
		id,
	})
	if err != nil {
		return nil, err
	}

	// btree index on (name, id)
	if err := table.AddIndex(indexes.NewBTreeIndex(
		"k1_name_id",
		table.Name(),
		[]expressions.ExpressionWithDirection{
			{Expression: expressions.NewNamed(name)},
			{Expression: expressions.NewNamed(id)},
		},
	)); err != nil {
		return nil, err
	}

	return table, nil
}

func createK2Table(root string) (*table.Table, error) {
	name := shared.NewField("k2", "name", shared.TypeKindText)
	id := shared.NewField("k2", "id", shared.TypeKindNumeric)

	table, err := loader.NewTableFromCSV("k2", csvFilepath(root, "k2"), []shared.Field{
		name,
		id,
	})
	if err != nil {
		return nil, err
	}

	// btree index on (name, id)
	if err := table.AddIndex(indexes.NewBTreeIndex(
		"k2_name_id",
		table.Name(),
		[]expressions.ExpressionWithDirection{
			{Expression: expressions.NewNamed(name)},
			{Expression: expressions.NewNamed(id)},
		},
	)); err != nil {
		return nil, err
	}

	return table, nil
}

func csvFilepath(root, name string) string {
	return filepath.Join(root, fmt.Sprintf("testdata/%s.csv", name))
}
