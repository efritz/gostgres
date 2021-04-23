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
	return copyFields(n.projector.fields)
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

func (n *updateNode) Ordering() OrderExpression {
	return nil
}

func (n *updateNode) Scan(visitor VisitorFunc) error {
	return n.Node.Scan(n.decorateVisitor(visitor))
}

func (n *updateNode) decorateVisitor(visitor VisitorFunc) VisitorFunc {
	return func(row shared.Row) (bool, error) {
		values := make([]interface{}, len(row.Values))
		copy(values, row.Values)

		for _, set := range n.setExpressions {
			value, err := set.Expression.ValueFrom(row)
			if err != nil {
				return false, err
			}

			found := false
			for i, field := range row.Fields {
				if field.Name == set.Name {
					found = true
					values[i] = value
				}
			}

			if !found {
				return false, fmt.Errorf("unknown column %s", set.Name)
			}
		}

		updatedRow, err := shared.NewRow(row.Fields, values)
		if err != nil {
			return false, err
		}

		ok, err := n.table.Update(updatedRow)
		if err != nil {
			return false, err
		}
		if !ok {
			return true, nil
		}

		if len(n.projector.aliases) == 0 {
			return true, nil
		}

		projectedRow, err := n.projector.projectRow(updatedRow)
		if err != nil {
			return false, err
		}

		return visitor(projectedRow)
	}
}
