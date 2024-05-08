package ddl

import (
	"github.com/efritz/gostgres/internal/protocol"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/table"
)

type createTable struct {
	name    string
	columns []table.ColumnDefinition
}

var _ queries.Query = &createTable{}

func NewCreateTable(name string, columns []table.ColumnDefinition) *createTable {
	return &createTable{
		name:    name,
		columns: columns,
	}
}

func (n *createTable) Execute(ctx queries.Context, w protocol.ResponseWriter) {
	if err := n.execute(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (n *createTable) execute(ctx queries.Context) error {
	if err := ctx.Tables.CreateTable(n.name, n.columns); err != nil {
		return err
	}

	return nil
}
