package ddl

import (
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ddlSet struct {
	queries []DDLQuery
}

type DDLQuery interface {
	queries.Query
	ExecuteDDL(ctx impls.ExecutionContext) error
}

func NewSet(queries []DDLQuery) queries.Query {
	return &ddlSet{
		queries: queries,
	}
}

func (q *ddlSet) Execute(ctx impls.ExecutionContext, w protocol.ResponseWriter) {
	for _, query := range q.queries {
		if err := query.ExecuteDDL(ctx); err != nil {
			w.Error(err)
			return
		}
	}

	w.Done()
}
