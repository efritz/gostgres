package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type insertNode struct {
	Node
	table       *Table
	columnNames []string
	projector   *projector
}

var _ Node = &insertNode{}

func NewInsert(node Node, table *Table, name, alias string, columnNames []string, expressions []ProjectionExpression) (Node, error) {
	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(name, table.Fields(), alias)
		}
	}

	projector, err := newProjector(node.Name(), table.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &insertNode{
		Node:        node,
		table:       table,
		columnNames: columnNames,
		projector:   projector,
	}, nil
}

func (r *insertNode) Fields() []shared.Field {
	return copyFields(r.projector.fields)
}

func (r *insertNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sinsert returning (%s)\n", indent(indentationLevel), r.projector))
	r.Node.Serialize(w, indentationLevel+1)
}

func (r *insertNode) Optimize() {
	r.projector.optimize()
	r.Node.Optimize()
}

func (r *insertNode) PushDownFilter(filter expressions.Expression) bool {
	return false
}

func (r *insertNode) Scan(visitor VisitorFunc) error {
	return r.Node.Scan(r.decorateVisitor(visitor))
}

func (r *insertNode) decorateVisitor(visitor VisitorFunc) VisitorFunc {
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

func (r *insertNode) prepareValuesForRow(row shared.Row) []interface{} {
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
