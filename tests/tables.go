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
	employeesTable, err := createEmployeesTable(root)
	if err != nil {
		return nil, err
	}

	departmentsTable, err := createDepartmentsTable(root)
	if err != nil {
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

	k1, err := createK1Table(root)
	if err != nil {
		return nil, err
	}

	k2, err := createK2Table(root)
	if err != nil {
		return nil, err
	}

	return map[string]*table.Table{
		"employees":   employeesTable,
		"departments": departmentsTable,
		"locations":   locationsTable,
		"regions":     regionsTable,
		"k1":          k1,
		"k2":          k2,
	}, nil
}

func createEmployeesTable(root string) (*table.Table, error) {
	table, err := loader.NewTableFromCSV("employees", csvFilepath(root, "employees"), []loader.FieldDescription{
		{Name: "employee_id", TypeKind: shared.TypeKindNumeric},
		{Name: "first_name", TypeKind: shared.TypeKindText},
		{Name: "last_name", TypeKind: shared.TypeKindText},
		{Name: "email", TypeKind: shared.TypeKindText},
		{Name: "manager_id", TypeKind: shared.TypeKindNumeric},
		{Name: "department_id", TypeKind: shared.TypeKindNumeric},
	})
	if err != nil {
		return nil, err
	}

	// btree index on (last_name, first_name, employee_id)
	if err := table.AddIndex(indexes.NewBTreeIndex("employees_last_name_first_name_employee_id", table.Name(), []expressions.ExpressionWithDirection{
		{Expression: expressions.NewNamed(shared.NewField("employees", "last_name", shared.TypeKindText, false))},
		{Expression: expressions.NewNamed(shared.NewField("employees", "first_name", shared.TypeKindText, false))},
		{Expression: expressions.NewNamed(shared.NewField("employees", "employee_id", shared.TypeKindNumeric, false))},
	})); err != nil {
		return nil, err
	}

	// hash index on first name
	if err := table.AddIndex(indexes.NewHashIndex("employees_first_name", table.Name(),
		expressions.NewNamed(shared.NewField("employees", "first_name", shared.TypeKindText, false)),
	)); err != nil {
		return nil, err
	}

	// hash index last_name, partial where manager_id <= 4
	lastName := expressions.NewNamed(shared.NewField("employees", "last_name", shared.TypeKindText, false))
	manager := expressions.NewNamed(shared.NewField("employees", "manager_id", shared.TypeKindNumeric, false))
	index := indexes.NewHashIndex("employees_last_name_manager_id", table.Name(), lastName)
	cond := expressions.NewLessThanEquals(manager, expressions.NewConstant(4))
	if err := table.AddIndex(indexes.NewPartialIndex(index, cond)); err != nil {
		return nil, err
	}

	return table, nil
}

func createDepartmentsTable(root string) (*table.Table, error) {
	table, err := loader.NewTableFromCSV("departments", csvFilepath(root, "departments"), []loader.FieldDescription{
		{Name: "department_id", TypeKind: shared.TypeKindNumeric},
		{Name: "department_name", TypeKind: shared.TypeKindText},
		{Name: "location_id", TypeKind: shared.TypeKindNumeric},
	})
	if err != nil {
		return nil, err
	}

	// hash index on department_id
	if err := table.AddIndex(indexes.NewHashIndex("departments_department_id", table.Name(),
		expressions.NewNamed(shared.NewField("departments", "department_id", shared.TypeKindNumeric, false)),
	)); err != nil {
		return nil, err
	}

	return table, nil
}

func createLocationsTable(root string) (*table.Table, error) {
	return loader.NewTableFromCSV("locations", csvFilepath(root, "locations"), []loader.FieldDescription{
		{Name: "location_id", TypeKind: shared.TypeKindNumeric},
		{Name: "location_name", TypeKind: shared.TypeKindText},
		{Name: "region_id", TypeKind: shared.TypeKindNumeric},
	})
}

func createRegionsTable(root string) (*table.Table, error) {
	return loader.NewTableFromCSV("regions", csvFilepath(root, "regions"), []loader.FieldDescription{
		{Name: "region_id", TypeKind: shared.TypeKindNumeric},
		{Name: "region_name", TypeKind: shared.TypeKindText},
	})
}

func createK1Table(root string) (*table.Table, error) {
	table, err := loader.NewTableFromCSV("k1", csvFilepath(root, "k1"), []loader.FieldDescription{
		{Name: "name", TypeKind: shared.TypeKindText},
		{Name: "id", TypeKind: shared.TypeKindNumeric},
	})
	if err != nil {
		return nil, err
	}

	// btree index on (name, id)
	if err := table.AddIndex(indexes.NewBTreeIndex("k1_name_id", table.Name(), []expressions.ExpressionWithDirection{
		{Expression: expressions.NewNamed(shared.NewField("k1", "name", shared.TypeKindText, false))},
		{Expression: expressions.NewNamed(shared.NewField("k1", "id", shared.TypeKindNumeric, false))},
	})); err != nil {
		return nil, err
	}

	return table, nil
}

func createK2Table(root string) (*table.Table, error) {
	table, err := loader.NewTableFromCSV("k2", csvFilepath(root, "k2"), []loader.FieldDescription{
		{Name: "name", TypeKind: shared.TypeKindText},
		{Name: "id", TypeKind: shared.TypeKindNumeric},
	})
	if err != nil {
		return nil, err
	}

	// btree index on (name, id)
	if err := table.AddIndex(indexes.NewBTreeIndex("k2_name_id", table.Name(), []expressions.ExpressionWithDirection{
		{Expression: expressions.NewNamed(shared.NewField("k2", "name", shared.TypeKindText, false))},
		{Expression: expressions.NewNamed(shared.NewField("k2", "id", shared.TypeKindNumeric, false))},
	})); err != nil {
		return nil, err
	}

	return table, nil
}

func csvFilepath(root, name string) string {
	return filepath.Join(root, fmt.Sprintf("testdata/%s.csv", name))
}
