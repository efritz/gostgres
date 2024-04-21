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

type Tablespace struct {
	tables map[string]*table.Table
}

func NewTablespace() *Tablespace {
	return &Tablespace{
		tables: map[string]*table.Table{},
	}
}

func (t *Tablespace) GetTable(name string) (*table.Table, bool) {
	table, ok := t.tables[name]
	return table, ok
}

func (t *Tablespace) CreateTable(name string, fields []shared.Field) error {
	_, err := t.CreateAndGetTable(name, fields)
	return err
}

func (t *Tablespace) CreateAndGetTable(name string, fields []shared.Field) (*table.Table, error) {
	table := table.NewTable(name, fields)
	t.tables[name] = table
	return table, nil
}

func CreateStandardTestTables(root string) (*Tablespace, error) {
	loaders := []func(*Tablespace, string) error{
		createEmployeesTable,
		createDepartmentsTable,
		createLocationsTable,
		createRegionsTable,
		createK1Table,
		createK2Table,
	}

	tables := NewTablespace()
	for _, loader := range loaders {
		if err := loader(tables, root); err != nil {
			return nil, err
		}
	}

	return tables, nil
}

func createEmployeesTable(tables *Tablespace, root string) error {
	employeeID := shared.NewField("employees", "employee_id", shared.TypeNumeric)
	firstName := shared.NewField("employees", "first_name", shared.TypeText)
	last_name := shared.NewField("employees", "last_name", shared.TypeText)
	email := shared.NewField("employees", "email", shared.TypeText)
	managerID := shared.NewField("employees", "manager_id", shared.TypeNumeric)
	departmentID := shared.NewField("employees", "department_id", shared.TypeNumeric)
	bonus := shared.NewField("employees", "bonus", shared.TypeNullableNumeric)
	fields := []shared.Field{
		employeeID,
		firstName,
		last_name,
		email,
		managerID,
		departmentID,
		bonus,
	}

	table, err := tables.CreateAndGetTable("employees", fields)
	if err != nil {
		return err
	}
	if err := loader.PopulateTableFromCSV(table, csvFilepath(root, "employees")); err != nil {
		return err
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
		return err
	}

	// hash index on first name
	if err := table.AddIndex(indexes.NewHashIndex(
		"employees_first_name",
		table.Name(),
		expressions.NewNamed(firstName),
	)); err != nil {
		return err
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
		return err
	}

	return nil
}

func createDepartmentsTable(tables *Tablespace, root string) error {
	departmentID := shared.NewField("departments", "department_id", shared.TypeNumeric)
	departmentName := shared.NewField("departments", "department_name", shared.TypeText)
	locationID := shared.NewField("departments", "location_id", shared.TypeNumeric)
	fields := []shared.Field{
		departmentID,
		departmentName,
		locationID,
	}

	table, err := tables.CreateAndGetTable("departments", fields)
	if err != nil {
		return err
	}
	if err := loader.PopulateTableFromCSV(table, csvFilepath(root, "departments")); err != nil {
		return err
	}

	// hash index on department_id
	if err := table.AddIndex(indexes.NewHashIndex(
		"departments_department_id",
		table.Name(),
		expressions.NewNamed(departmentID),
	)); err != nil {
		return err
	}

	return nil
}

func createLocationsTable(tables *Tablespace, root string) error {
	locationID := shared.NewField("locations", "location_id", shared.TypeNumeric)
	locationName := shared.NewField("locations", "location_name", shared.TypeText)
	regionID := shared.NewField("locations", "region_id", shared.TypeNumeric)
	fields := []shared.Field{
		locationID,
		locationName,
		regionID,
	}

	table, err := tables.CreateAndGetTable("locations", fields)
	if err != nil {
		return err
	}
	if err := loader.PopulateTableFromCSV(table, csvFilepath(root, "locations")); err != nil {
		return err
	}

	return nil
}

func createRegionsTable(tables *Tablespace, root string) error {
	regionID := shared.NewField("regions", "region_id", shared.TypeNumeric)
	regionName := shared.NewField("regions", "region_name", shared.TypeText)
	fields := []shared.Field{
		regionID,
		regionName,
	}

	table, err := tables.CreateAndGetTable("regions", fields)
	if err != nil {
		return err
	}
	if err := loader.PopulateTableFromCSV(table, csvFilepath(root, "regions")); err != nil {
		return err
	}

	return nil
}

func createK1Table(tables *Tablespace, root string) error {
	name := shared.NewField("k1", "name", shared.TypeText)
	id := shared.NewField("k1", "id", shared.TypeNumeric)
	fields := []shared.Field{
		name,
		id,
	}

	table, err := tables.CreateAndGetTable("k1", fields)
	if err != nil {
		return err
	}
	if err := loader.PopulateTableFromCSV(table, csvFilepath(root, "k1")); err != nil {
		return err
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
		return err
	}

	return nil
}

func createK2Table(tables *Tablespace, root string) error {
	name := shared.NewField("k2", "name", shared.TypeText)
	id := shared.NewField("k2", "id", shared.TypeNumeric)
	fields := []shared.Field{
		name,
		id,
	}

	table, err := tables.CreateAndGetTable("k2", fields)
	if err != nil {
		return err
	}
	if err := loader.PopulateTableFromCSV(table, csvFilepath(root, "k2")); err != nil {
		return err
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
		return err
	}

	return nil
}

func csvFilepath(root, name string) string {
	return filepath.Join(root, fmt.Sprintf("testdata/%s.csv", name))
}
