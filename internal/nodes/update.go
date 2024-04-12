package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type updateNode struct {
	Node
	table          *Table
	setExpressions []SetExpression
	columnNames    []string
	projector      *projector
}

var _ Node = &updateNode{}

type SetExpression struct {
	Name       string
	Expression expressions.Expression
}

func NewUpdate(node Node, table *Table, setExpressions []SetExpression, alias string, expressions []ProjectionExpression) (Node, error) {
	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(table.name, table.Fields(), alias)
		}
	}

	projector, err := newProjector(node.Name(), table.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &updateNode{
		Node:           node,
		table:          table,
		setExpressions: setExpressions,
		projector:      projector,
	}, nil
}

func (n *updateNode) Fields() []shared.Field {
	return copyFields(n.projector.projectedFields)
}

func (n *updateNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%supdate returning (%s)\n", indent(indentationLevel), n.projector))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *updateNode) Optimize() {
	n.projector.optimize()
	n.Node.Optimize()
}

func (n *updateNode) AddFilter(filter expressions.Expression) {
}

func (n *updateNode) AddOrder(order OrderExpression) {
}

func (n *updateNode) Filter() expressions.Expression {
	return nil
}

func (n *updateNode) Ordering() OrderExpression {
	return nil
}

func (n *updateNode) SupportsMarkRestore() bool {
	return false
}

func (n *updateNode) Scanner() (Scanner, error) {
	scanner, err := n.Node.Scanner()
	if err != nil {
		return nil, err
	}

	return ScannerFunc(func() (shared.Row, error) {
		row, err := scanner.Scan()
		if err != nil {
			return shared.Row{}, err
		}

		values := make([]interface{}, len(row.Values))
		copy(values, row.Values)

		for _, set := range n.setExpressions {
			value, err := set.Expression.ValueFrom(row)
			if err != nil {
				return shared.Row{}, err
			}

			found := false
			for i, field := range row.Fields {
				if field.Name == set.Name {
					if field.Internal {
						return shared.Row{}, fmt.Errorf("cannot update internal field %s", set.Name)
					}

					found = true
					values[i] = value
				}
			}

			if !found {
				return shared.Row{}, fmt.Errorf("unknown column %s", set.Name)
			}
		}

		deletedRow, err := shared.NewRow(row.Fields[:1], values[:1])
		if err != nil {
			return shared.Row{}, err
		}
		if _, ok, err := n.table.Delete(deletedRow); err != nil {
			return shared.Row{}, err
		} else if !ok {
			return shared.Row{}, nil
		}

		insertedRow, err := shared.NewRow(row.Fields[1:], values[1:])
		if err != nil {
			return shared.Row{}, err
		}

		updatedRow, err := n.table.Insert(insertedRow)
		if err != nil {
			return shared.Row{}, err
		}

		return n.projector.projectRow(updatedRow)
	}), nil
}
