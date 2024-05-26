package ddl

import (
	"github.com/efritz/gostgres/internal/catalog/sequence"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type createSequence struct {
	name string
	typ  shared.Type
}

var _ queries.Query = &createSequence{}
var _ DDLQuery = &createSequence{}

func NewCreateSequence(name string, typ shared.Type) *createSequence {
	return &createSequence{
		name: name,
		typ:  typ,
	}
}

func (q *createSequence) Execute(ctx types.Context, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createSequence) ExecuteDDL(ctx types.Context) error {
	ctx.SetSequence(q.name, sequence.NewSequence(q.name, q.typ))
	return nil
}
