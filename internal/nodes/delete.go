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

func (n *deleteNode) Scanner() (Scanner, error) {
	scanner, err := n.Node.Scanner()
	if err != nil {
		return nil, err
	}

	return ScannerFunc(func() (shared.Row, error) {
		for {
			row, err := scanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			deletedRow, ok, err := n.table.Delete(row)
			if err != nil {
				return shared.Row{}, err
			}
			if !ok {
				continue
			}

			// TODO - necessary?
			if len(n.projector.aliases) == 0 {
				return shared.Row{}, nil
			}

			return n.projector.projectRow(deletedRow)
		}
	}), nil
}
