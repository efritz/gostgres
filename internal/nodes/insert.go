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

func (n *insertNode) Fields() []shared.Field {
	return copyFields(n.projector.fields)
}

func (n *insertNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sinsert returning (%s)\n", indent(indentationLevel), n.projector))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *insertNode) Optimize() {
	n.projector.optimize()
	n.Node.Optimize()
}

func (n *insertNode) PushDownFilter(filter expressions.Expression) bool {
	return false
}

func (n *insertNode) Scan(visitor VisitorFunc) error {
	return n.Node.Scan(n.decorateVisitor(visitor))
}

func (n *insertNode) decorateVisitor(visitor VisitorFunc) VisitorFunc {
	return func(row shared.Row) (bool, error) {
		fields := make([]shared.Field, 0, len(n.table.Fields()))
		for _, field := range n.table.Fields() {
			if !field.Internal {
				fields = append(fields, field)
			}
		}

		insertedRow, err := shared.NewRow(fields, n.prepareValuesForRow(row))
		if err != nil {
			return false, err
		}

		insertedRow, err = n.table.Insert(insertedRow)
		if err != nil {
			return false, err
		}

		if len(n.projector.aliases) == 0 {
			return true, nil
		}

		projectedRow, err := n.projector.projectRow(insertedRow)
		if err != nil {
			return false, err
		}

		return visitor(projectedRow)
	}
}

func (n *insertNode) prepareValuesForRow(row shared.Row) []interface{} {
	if n.columnNames == nil {
		return row.Values
	}

	valueMap := make(map[string]interface{}, len(n.columnNames))
	for i, name := range n.columnNames {
		valueMap[name] = row.Values[i]
	}

	values := make([]interface{}, 0, len(n.table.Fields()))
	for _, field := range n.table.Fields() {
		values = append(values, valueMap[field.Name])
	}

	return values
}
