package tests

import (
	"fmt"
	"path/filepath"

	"github.com/efritz/gostgres/internal/loader"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/lexing"
	"github.com/efritz/gostgres/internal/syntax/parsing"
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
	statements := []string{
		`
			CREATE TABLE employees (
				employee_id integer NOT NULL,
				first_name text NOT NULL,
				last_name text NOT NULL,
				email text NOT NULL,
				manager_id integer NOT NULL,
				department_id integer NOT NULL,
				bonus integer
			)
		`,
		`CREATE INDEX employees_last_name_first_name_employee_id ON employees(last_name, first_name, employee_id)`,
		`CREATE INDEX employees_first_name ON employees USING hash(first_name)`,
		`CREATE INDEX employees_last_name_manager_id ON employees USING hash(last_name) WHERE manager_id <= 4`,

		`CREATE TABLE departments (department_id integer NOT NULL, department_name text NOT NULL, location_id integer NOT NULL)`,
		`CREATE INDEX departments_department_id ON departments USING hash(department_id)`,

		`CREATE TABLE locations (location_id integer NOT NULL, location_name text NOT NULL, region_id integer NOT NULL)`,
		`CREATE TABLE regions (region_id integer NOT NULL, region_name text NOT NULL)`,

		`CREATE TABLE k1 (name text NOT NULL, id integer NOT NULL)`,
		`CREATE INDEX k1_name_id ON k1 USING btree(name, id)`,

		`CREATE TABLE k2 (name text NOT NULL, id integer NOT NULL)`,
		`CREATE INDEX k2_name_id ON k2 USING btree(name, id)`,
	}

	tables := NewTablespace()

	for _, statement := range statements {
		node, err := parsing.Parse(lexing.Lex(statement), tables)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node: %s", err)
		}
		if _, err := node.Scanner(scan.ScanContext{Tables: tables}); err != nil {
			return nil, err
		}
	}

	tableNames := []string{
		"employees",
		"departments",
		"locations",
		"regions",
		"k1",
		"k2",
	}
	for _, name := range tableNames {
		table, ok := tables.GetTable(name)
		if !ok {
			return nil, fmt.Errorf("table %q not found", name)
		}
		if err := loader.PopulateTableFromCSV(table, csvFilepath(root, name)); err != nil {
			return nil, err
		}
	}

	return tables, nil
}

func csvFilepath(root, name string) string {
	return filepath.Join(root, fmt.Sprintf("testdata/%s.csv", name))
}
