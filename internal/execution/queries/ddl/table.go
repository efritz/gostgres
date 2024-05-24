package ddl

import (
	"github.com/efritz/gostgres/internal/catalog/table"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
)

type createTable struct {
	name   string
	fields []table.TableField
}

var _ queries.Query = &createTable{}
var _ DDLQuery = &createTable{}

func NewCreateTable(name string, fields []table.TableField) *createTable {
	return &createTable{
		name:   name,
		fields: fields,
	}
}

func (q *createTable) Execute(ctx queries.Context, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createTable) ExecuteDDL(ctx queries.Context) error {
	if err := ctx.CreateTable(q.name, q.fields); err != nil {
		return err
	}

	return nil
}
