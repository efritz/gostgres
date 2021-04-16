package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type joinNode struct {
	left   Node
	right  Node
	filter expressions.Expression
	fields []shared.Field
}

var _ Node = &joinNode{}

func NewJoin(left Node, right Node, condition expressions.Expression) Node {
	return &joinNode{
		left:   left,
		right:  right,
		filter: condition,
		fields: append(joinFieldsForNode(left), joinFieldsForNode(right)...),
	}
}

func (n *joinNode) Name() string {
	return ""
}

func (n *joinNode) Fields() []shared.Field {
	return copyFields(n.fields)
}

func (n *joinNode) Serialize(w io.Writer, indentationLevel int) {
	indentation := indent(indentationLevel)
	io.WriteString(w, fmt.Sprintf("%sjoin\n", indentation))
	n.left.Serialize(w, indentationLevel+1)
	io.WriteString(w, fmt.Sprintf("%swith\n", indentation))
	n.right.Serialize(w, indentationLevel+1)

	if n.filter != nil {
		io.WriteString(w, fmt.Sprintf("%son %s\n", indentation, n.filter))
	}
}

func (n *joinNode) Optimize() {
	if n.filter != nil {
		n.filter = n.distributeFilter(n.filter.Fold())
	}

	n.left.Optimize()
	n.right.Optimize()
}

func (n *joinNode) distributeFilter(filter expressions.Expression) expressions.Expression {
	var conjunctions []expressions.Expression
	for _, expression := range filter.Conjunctions() {
		if !n.distributeExpression(expression) {
			conjunctions = append(conjunctions, expression)
		}
	}

	return combineConjunctions(conjunctions)
}

func (n *joinNode) distributeExpression(expression expressions.Expression) bool {
	namesMatchingInLeft := false
	namesMissingFromLeft := false
	namesMatchingInRight := false
	namesMissingFromRight := false

	for _, field := range expression.Fields() {
		if _, err := shared.FindMatchingFieldIndex(field, n.left.Fields()); err != nil {
			namesMissingFromLeft = true
		} else {
			namesMatchingInLeft = true
		}
		if _, err := shared.FindMatchingFieldIndex(field, n.right.Fields()); err != nil {
			namesMissingFromRight = true
		} else {
			namesMatchingInRight = true
		}
	}

	pushedLeft := false
	if !namesMissingFromLeft {
		pushedLeft = n.left.PushDownFilter(expression)
	}

	pushedRight := false
	if !namesMissingFromRight {
		pushedRight = n.right.PushDownFilter(expression)
	}

	return (namesMatchingInLeft || !namesMissingFromLeft) == pushedLeft && (namesMatchingInRight || !namesMissingFromRight) == pushedRight
}

func (n *joinNode) PushDownFilter(filter expressions.Expression) bool {
	if n.filter != nil {
		filter = expressions.NewAnd(n.filter, filter)
	}

	n.filter = filter
	return true
}

func (n *joinNode) Scan(visitor VisitorFunc) error {
	return n.left.Scan(n.decorateLeftVisitor(visitor))
}

func (n *joinNode) decorateLeftVisitor(visitor VisitorFunc) VisitorFunc {
	return func(leftRow shared.Row) (bool, error) {
		return true, n.right.Scan(n.decorateRightVisitor(visitor, leftRow))
	}
}

func (n *joinNode) decorateRightVisitor(visitor VisitorFunc, leftRow shared.Row) VisitorFunc {
	return func(rightRow shared.Row) (bool, error) {
		row, err := shared.NewRow(n.Fields(), append(copyValues(leftRow.Values), rightRow.Values...))
		if err != nil {
			return false, err
		}

		if n.filter != nil {
			if ok, err := shared.EnsureBool(n.filter.ValueFrom(row)); err != nil {
				return false, err
			} else if !ok {
				return true, nil
			}
		}

		return visitor(row)
	}
}

func joinFieldsForNode(node Node) []shared.Field {
	if node.Name() == "" {
		return node.Fields()
	}

	return updateRelationName(node.Fields(), node.Name())
}
