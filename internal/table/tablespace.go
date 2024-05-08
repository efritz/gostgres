package table

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type Tablespace struct {
	tables map[string]*Table
}

func NewTablespace() *Tablespace {
	return &Tablespace{
		tables: map[string]*Table{},
	}
}

func (t *Tablespace) GetTable(name string) (*Table, bool) {
	table, ok := t.tables[name]
	return table, ok
}

type ColumnDefinition struct {
	Field       shared.Field
	Constraints []expressions.Expression
}

func (t *Tablespace) CreateTable(name string, columns []ColumnDefinition) error {
	_, err := t.CreateAndGetTable(name, columns)
	return err
}

func (t *Tablespace) CreateAndGetTable(name string, columns []ColumnDefinition) (*Table, error) {
	fields := []shared.Field{}
	for _, column := range columns {
		fields = append(fields, column.Field)
	}

	table := NewTable(name, fields)
	for _, column := range columns {
		for _, constraint := range column.Constraints {
			table.AddConstraint(NewConstraint(constraint))
		}
	}

	t.tables[name] = table
	return table, nil
}
