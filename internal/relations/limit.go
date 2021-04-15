package relations

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type limitRelation struct {
	Relation
	limit int
}

var _ Relation = &limitRelation{}

func NewLimit(relation Relation, limit int) Relation {
	return &limitRelation{
		Relation: relation,
		limit:    limit,
	}
}

func (r *limitRelation) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%slimit %d\n", indent(indentationLevel), r.limit))
	r.Relation.Serialize(w, indentationLevel+1)
}

func (r *limitRelation) Optimize() {
	r.Relation.Optimize()
}

func (r *limitRelation) PushDownFilter(filter expressions.Expression) bool {
	// filter boundary
	return false
}

func (r *limitRelation) Scan(visitor VisitorFunc) error {
	remaining := r.limit
	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		if remaining <= 0 {
			return false, nil
		}

		remaining--
		return visitor(row)
	})
}
