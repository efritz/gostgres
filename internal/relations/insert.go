package relations

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type insertRelation struct {
	Relation
	table       *Table
	columnNames []string
}

var _ Relation = &insertRelation{}

func NewInsert(relation Relation, table *Table, columnNames []string) Relation {
	return &insertRelation{
		Relation:    relation,
		table:       table,
		columnNames: columnNames,
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
		var fields []shared.Field
		for _, columnName := range r.columnNames {
			fields = append(fields, shared.Field{
				Name: columnName,
			})
		}

		var values []interface{}
		for _, field := range r.table.Fields() {
			var value interface{}
			for i, f2 := range fields {
				if f2.Name == field.Name {
					value = row.Values[i]
				}
			}

			values = append(values, value)
		}

		return true, r.table.Insert(shared.Row{
			Fields: r.table.Fields(),
			Values: values,
		})
	})
}
