package table

import (
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

func (t *Tablespace) CreateTable(name string, fields []shared.Field) error {
	_, err := t.CreateAndGetTable(name, fields)
	return err
}

func (t *Tablespace) CreateAndGetTable(name string, fields []shared.Field) (*Table, error) {
	table := NewTable(name, fields)
	t.tables[name] = table
	return table, nil
}
