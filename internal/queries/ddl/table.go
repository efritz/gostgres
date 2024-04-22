package ddl

import (
	"github.com/efritz/gostgres/internal/protocol"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/shared"
)

type createTable struct {
	name   string
	fields []shared.Field
}

var _ queries.Query = &createTable{}

func NewCreateTable(name string, fields []shared.Field) *createTable {
	return &createTable{
		name:   name,
		fields: fields,
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
	if err := ctx.Tables.CreateTable(n.name, n.fields); err != nil {
		return err
	}

	return nil
}
