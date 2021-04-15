package relations

import (
	"bytes"
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type limitRelation struct {
	Relation
	limit int
}

var _ Relation = &limitRelation{}

func NewLimit(table Relation, limit int) Relation {
	return &limitRelation{
		Relation: table,
		limit:    limit,
	}
}

func (r *limitRelation) Serialize(buf *bytes.Buffer, indentationLevel int) {
	buf.WriteString(fmt.Sprintf("%slimit %d\n", indent(indentationLevel), r.limit))
	r.Relation.Serialize(buf, indentationLevel+1)
}

func (r *limitRelation) Optimize() {
	r.Relation.Optimize()
}

func (r *limitRelation) PushDownFilter(filter expressions.Expression) {
	// filter boundary
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
