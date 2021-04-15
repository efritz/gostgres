package relations

import (
	"bytes"
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
)

type orderRelation struct {
	Relation
	order expressions.Expression
}

var _ Relation = &orderRelation{}

func NewOrder(relation Relation, order expressions.Expression) Relation {
	return &orderRelation{
		Relation: relation,
		order:    order,
	}
}

func (r *orderRelation) Serialize(buf *bytes.Buffer, indentationLevel int) {
	if r.order == nil {
		r.Relation.Serialize(buf, indentationLevel)
		return
	}

	buf.WriteString(fmt.Sprintf("%sorder by %s\n", indent(indentationLevel), r.order))
	r.Relation.Serialize(buf, indentationLevel+1)
}

func (r *orderRelation) Optimize() {
	if r.order != nil {
		r.order = r.order.Fold()
	}

	r.Relation.Optimize()
}

func (r *orderRelation) PushDownFilter(filter expressions.Expression) {
	r.Relation.PushDownFilter(filter)
}

func (r *orderRelation) Scan(visitor VisitorFunc) error {
	if r.order == nil {
		return r.Relation.Scan(visitor)
	}

	rows, err := ScanRows(r.Relation)
	if err != nil {
		return err
	}

	indexes, err := findIndexIterationOrder(r.order, rows)
	if err != nil {
		return err
	}

	for _, i := range indexes {
		if ok, err := visitor(rows.Row(i)); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
