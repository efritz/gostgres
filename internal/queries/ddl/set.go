package ddl

import (
	"github.com/efritz/gostgres/internal/protocol"
	"github.com/efritz/gostgres/internal/queries"
)

type ddlSet struct {
	queries []DDLQuery
}

type DDLQuery interface {
	queries.Query
	ExecuteDDL(ctx queries.Context) error
}

var _ queries.Query = &ddlSet{}

func NewSet(queries []DDLQuery) *ddlSet {
	return &ddlSet{
		queries: queries,
	}
}

func (q *ddlSet) Execute(ctx queries.Context, w protocol.ResponseWriter) {
	for _, query := range q.queries {
		if err := query.ExecuteDDL(ctx); err != nil {
			w.Error(err)
			return
		}
	}

	w.Done()
}