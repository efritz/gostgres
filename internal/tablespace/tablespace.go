package tablespace

import (
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
