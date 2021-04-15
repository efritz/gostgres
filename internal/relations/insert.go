package relations

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type insertRelation struct {
	Relation
	table *Table
}

var _ Relation = &insertRelation{}

func NewInsert(relation Relation, table *Table) Relation {
	return &insertRelation{
		Relation: relation,
		table:    table,
	}
}

func (r *insertRelation) Fields() []shared.Field {
	return nil
}

func (r *insertRelation) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sinsert\n", indent(indentationLevel)))
	r.Relation.Serialize(w, indentationLevel+1)
}

func (r *insertRelation) Optimize() {
	r.Relation.Optimize()
}

func (r *insertRelation) PushDownFilter(filter expressions.Expression) bool {
	return false
}

func (r *insertRelation) Scan(visitor VisitorFunc) error {
	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		return true, r.table.Insert(row)
	})
}
