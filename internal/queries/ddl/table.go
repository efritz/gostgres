package ddl

import (
	"github.com/efritz/gostgres/internal/protocol"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/shared"
)

type createTable struct {
	name   string
	fields []shared.TableField
}

var _ queries.Query = &createTable{}
var _ DDLQuery = &createTable{}

func NewCreateTable(name string, fields []shared.TableField) *createTable {
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
	if err := ctx.Tables.CreateTable(q.name, q.fields); err != nil {
		return err
	}

	return nil
}
