package ddl

import (
	"github.com/efritz/gostgres/internal/catalog/table"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type createTable struct {
	name   string
	fields []impls.TableField
}

var _ queries.Query = &createTable{}
var _ DDLQuery = &createTable{}

func NewCreateTable(name string, fields []impls.TableField) *createTable {
	return &createTable{
		name:   name,
		fields: fields,
	}
}

func (q *createTable) Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createTable) ExecuteDDL(ctx impls.ExecutionContext) error {
	ctx.Catalog().Tables.Set(q.name, table.NewTable(q.name, q.fields))
	return nil
}
