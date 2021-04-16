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
	projector   *projector
}

var _ Relation = &insertRelation{}

func NewInsert(relation Relation, table *Table, name, alias string, columnNames []string, expressions []ProjectionExpression) (Relation, error) {
	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(name, table.Fields(), alias)
		}
	}

	projector, err := newProjector(relation.Name(), table.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &insertRelation{
		Relation:    relation,
		table:       table,
		columnNames: columnNames,
		projector:   projector,
	}, nil
}

func (r *insertRelation) Fields() []shared.Field {
	return copyFields(r.projector.fields)
}

func (r *insertRelation) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sinsert returning (%s)\n", indent(indentationLevel), r.projector))
	r.Relation.Serialize(w, indentationLevel+1)
}

func (r *insertRelation) Optimize() {
	r.projector.optimize()
	r.Relation.Optimize()
}

func (r *insertRelation) PushDownFilter(filter expressions.Expression) bool {
	return false
}

func (r *insertRelation) Scan(visitor VisitorFunc) error {
	return r.Relation.Scan(func(row shared.Row) (bool, error) {
		var values []interface{}
		if r.columnNames == nil {
			values = row.Values
		} else {
			var fields []shared.Field
			for _, columnName := range r.columnNames {
				fields = append(fields, shared.Field{
					Name: columnName,
				})
			}

			for _, f1 := range r.table.Fields() {
				var value interface{}
				for i, f2 := range fields {
					if f1.Name == f2.Name {
						value = row.Values[i]
					}
				}

				values = append(values, value)
			}
		}

		insertedRow := shared.NewRow(r.table.Fields(), values)
		if err := r.table.Insert(insertedRow); err != nil {
			return false, err
		}

		if len(r.projector.aliases) == 0 {
			return true, nil
		}

		projectedRow, err := r.projector.projectRow(insertedRow)
		if err != nil {
			return false, err
		}

		return visitor(projectedRow)
	})
}
