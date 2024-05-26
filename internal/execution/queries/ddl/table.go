package ddl

import (
	"github.com/efritz/gostgres/internal/catalog/table"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/types"
)

type createTable struct {
	name   string
	fields []types.TableField
}

var _ queries.Query = &createTable{}
var _ DDLQuery = &createTable{}

func NewCreateTable(name string, fields []types.TableField) *createTable {
	return &createTable{
		name:   name,
		fields: fields,
	}
}

func (q *createTable) Execute(ctx types.Context, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createTable) ExecuteDDL(ctx types.Context) error {
	ctx.SetTable(q.name, table.NewTable(q.name, q.fields))
	return nil
}
