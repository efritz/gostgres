package relations

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type offsetRelation struct {
	Relation
	offset int
}

var _ Relation = &offsetRelation{}

func NewOffset(relation Relation, offset int) Relation {
	return &offsetRelation{
		Relation: relation,
		offset:   offset,
	}
}

func (r *offsetRelation) Serialize(w io.Writer, indentationLevel int) {
	if r.offset == 0 {
		r.Relation.Serialize(w, indentationLevel)
		return
	}

	io.WriteString(w, fmt.Sprintf("%soffset %d\n", indent(indentationLevel), r.offset))
	r.Relation.Serialize(w, indentationLevel+1)
}

func (r *offsetRelation) Optimize() {
	r.Relation.Optimize()
}

func (r *offsetRelation) PushDownFilter(filter expressions.Expression) bool {
	// filter boundary
	return false
}

func (r *offsetRelation) Scan(visitor VisitorFunc) error {
	if r.offset == 0 {
		return r.Relation.Scan(visitor)
	}

	offset := r.offset
	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		offset--
		if offset >= 0 {
			return true, nil
		}

		return visitor(row)
	})
}
