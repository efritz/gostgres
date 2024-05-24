package ddl

import (
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared"
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

func (q *createSequence) Execute(ctx queries.Context, w protocol.ResponseWriter) {
	if err := q.ExecuteDDL(ctx); err != nil {
		w.Error(err)
		return
	}

	w.Done()
}

func (q *createSequence) ExecuteDDL(ctx queries.Context) error {
	if _, err := ctx.CreateAndGetSequence(q.name, q.typ); err != nil {
		return err
	}

	return nil
}
