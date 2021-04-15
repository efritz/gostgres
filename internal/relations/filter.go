package relations

import (
	"bytes"
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type filterRelation struct {
	Relation
	filter expressions.Expression
}

var _ Relation = &filterRelation{}

func NewFilter(table Relation, filter expressions.Expression) Relation {
	return &filterRelation{
		Relation: table,
		filter:   filter,
	}
}

func (r *filterRelation) Serialize(buf *bytes.Buffer, indentationLevel int) {
	if r.filter == nil {
		r.Relation.Serialize(buf, indentationLevel)
		return
	}

	buf.WriteString(fmt.Sprintf("%sfilter by %s\n", indent(indentationLevel), r.filter))
	r.Relation.Serialize(buf, indentationLevel+1)
}

func (r *filterRelation) Optimize() {
	if r.filter != nil {
		r.filter = r.filter.Fold()
		r.PushDownFilter(r.filter)
	}

	r.Relation.Optimize()
}

func (r *filterRelation) PushDownFilter(filter expressions.Expression) {
	r.Relation.PushDownFilter(filter)
}

func (r *filterRelation) Scan(visitor VisitorFunc) error {
	if r.filter == nil {
		return r.Relation.Scan(visitor)
	}

	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		if ok, err := expressions.EnsureBool(r.filter.ValueFrom(row)); err != nil {
			return false, err
		} else if !ok {
			return true, nil
		}

		return visitor(row)
	})
}
