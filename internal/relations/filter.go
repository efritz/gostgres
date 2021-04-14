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
	r.Relation.Optimize()

	if r.filter != nil && r.Relation.SinkFilter(r.filter) {
		r.filter = nil
	}
}

func (r *filterRelation) SinkFilter(filter expressions.Expression) bool {
	return r.Relation.SinkFilter(filter)
}

func (r *filterRelation) Scan(visitor VisitorFunc) error {
	if r.filter == nil {
		return r.Relation.Scan(visitor)
	}

	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		if ok, err := expressions.Bool(r.filter).ValueFrom(row); err != nil {
			return false, err
		} else if !ok {
			return true, nil
		}

		return visitor(row)
	})
}
