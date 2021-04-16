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
	return r.Relation.Scan(r.decorateVisitor(visitor))
}

func (r *insertRelation) decorateVisitor(visitor VisitorFunc) VisitorFunc {
	return func(row shared.Row) (bool, error) {
		insertedRow := shared.NewRow(r.table.Fields(), r.prepareValuesForRow(row))
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
	}
}

func (r *insertRelation) prepareValuesForRow(row shared.Row) []interface{} {
	if r.columnNames == nil {
		return row.Values
	}

	valueMap := make(map[string]interface{}, len(r.columnNames))
	for i, name := range r.columnNames {
		valueMap[name] = row.Values[i]
	}

	values := make([]interface{}, 0, len(r.table.Fields()))
	for _, field := range r.table.Fields() {
		values = append(values, valueMap[field.Name])
	}

	return values
}
