package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type deleteNode struct {
	Node
	table       *Table
	columnNames []string
	projector   *projector
}

var _ Node = &deleteNode{}

func NewDelete(node Node, table *Table, alias string, expressions []ProjectionExpression) (Node, error) {
	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(table.name, table.Fields(), alias)
		}
	}

	projector, err := newProjector(node.Name(), table.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &deleteNode{
		Node:      node,
		table:     table,
		projector: projector,
	}, nil
}

func (n *deleteNode) Fields() []shared.Field {
	return copyFields(n.projector.projectedFields)
}

func (n *deleteNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sdelete returning (%s)\n", indent(indentationLevel), n.projector))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *deleteNode) Optimize() {
	n.projector.optimize()
	n.Node.Optimize()
}

func (n *deleteNode) AddFilter(filter expressions.Expression) {
}

func (n *deleteNode) AddOrder(order OrderExpression) {
}

func (n *deleteNode) Filter() expressions.Expression {
	return nil
}

func (n *deleteNode) Ordering() OrderExpression {
	return nil
}

func (n *deleteNode) Scan(visitor VisitorFunc) error {
	return n.Node.Scan(n.decorateVisitor(visitor))
}

func (n *deleteNode) decorateVisitor(visitor VisitorFunc) VisitorFunc {
	return func(row shared.Row) (bool, error) {
		deletedRow, ok, err := n.table.Delete(row)
		if err != nil {
			return false, err
		}
		if !ok {
			return true, nil
		}

		if len(n.projector.aliases) == 0 {
			return true, nil
		}

		projectedRow, err := n.projector.projectRow(deletedRow)
		if err != nil {
			return false, err
		}

		return visitor(projectedRow)
	}
}
