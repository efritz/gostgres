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
	buf.WriteString(fmt.Sprintf("%sorder by %s\n", indent(indentationLevel), r.order))
	r.Relation.Serialize(buf, indentationLevel+1)
}

func (r *orderRelation) Scan(visitor VisitorFunc) error {
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
