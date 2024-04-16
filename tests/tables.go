package tests

import (
	"fmt"
	"path/filepath"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/loader"
	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/shared"
)

func CreateStandardTestTables(root string) (map[string]*nodes.Table, error) {
	employeesTable, err := createEmployeesTable(root)
	if err != nil {
		return nil, err
	}
	if err := employeesTable.AddIndex(nodes.NewBTreeIndex("employees_last_name_first_name_employee_id", employeesTable, []nodes.ExpressionWithDirection{
		{Expression: expressions.NewNamed(shared.NewField("employees", "last_name", shared.TypeKindText, false))},
		{Expression: expressions.NewNamed(shared.NewField("employees", "first_name", shared.TypeKindText, false))},
		{Expression: expressions.NewNamed(shared.NewField("employees", "employee_id", shared.TypeKindNumeric, false))},
	})); err != nil {
		return nil, err
	}
	if err := employeesTable.AddIndex(nodes.NewHashIndex("employees_first_name", employeesTable,
		expressions.NewNamed(shared.NewField("employees", "first_name", shared.TypeKindText, false)),
	)); err != nil {
		return nil, err
	}

	departmentsTable, err := createDepartmentsTable(root)
	if err != nil {
		return nil, err
	}
	if err := departmentsTable.AddIndex(nodes.NewHashIndex("departments_department_id", departmentsTable,
		expressions.NewNamed(shared.NewField("departments", "department_id", shared.TypeKindNumeric, false)),
	)); err != nil {
		return nil, err
	}

	locationsTable, err := createLocationsTable(root)
	if err != nil {
		return nil, err
	}

	regionsTable, err := createRegionsTable(root)
	if err != nil {
		return nil, err
	}

	return map[string]*nodes.Table{
		"employees":   employeesTable,
		"departments": departmentsTable,
		"locations":   locationsTable,
		"regions":     regionsTable,
	}, nil
}

func createEmployeesTable(root string) (*nodes.Table, error) {
	return loader.NewTableFromCSV("employees", csvFilepath(root, "employees"), []loader.FieldDescription{
		{Name: "employee_id", TypeKind: shared.TypeKindNumeric},
		{Name: "first_name", TypeKind: shared.TypeKindText},
		{Name: "last_name", TypeKind: shared.TypeKindText},
		{Name: "email", TypeKind: shared.TypeKindText},
		{Name: "manager_id", TypeKind: shared.TypeKindNumeric},
		{Name: "department_id", TypeKind: shared.TypeKindNumeric},
	})
}

func createDepartmentsTable(root string) (*nodes.Table, error) {
	return loader.NewTableFromCSV("departments", csvFilepath(root, "departments"), []loader.FieldDescription{
		{Name: "department_id", TypeKind: shared.TypeKindNumeric},
		{Name: "department_name", TypeKind: shared.TypeKindText},
		{Name: "location_id", TypeKind: shared.TypeKindNumeric},
	})
}

func createLocationsTable(root string) (*nodes.Table, error) {
	return loader.NewTableFromCSV("locations", csvFilepath(root, "locations"), []loader.FieldDescription{
		{Name: "location_id", TypeKind: shared.TypeKindNumeric},
		{Name: "location_name", TypeKind: shared.TypeKindText},
		{Name: "region_id", TypeKind: shared.TypeKindNumeric},
	})
}

func createRegionsTable(root string) (*nodes.Table, error) {
	return loader.NewTableFromCSV("regions", csvFilepath(root, "regions"), []loader.FieldDescription{
		{Name: "region_id", TypeKind: shared.TypeKindNumeric},
		{Name: "region_name", TypeKind: shared.TypeKindText},
	})
}

func csvFilepath(root, name string) string {
	return filepath.Join(root, fmt.Sprintf("testdata/%s.csv", name))
}
