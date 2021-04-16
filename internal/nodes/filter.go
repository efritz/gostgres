package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type filterNode struct {
	Node
	filter expressions.Expression
}

var _ Node = &filterNode{}

func NewFilter(node Node, filter expressions.Expression) Node {
	return &filterNode{
		Node:   node,
		filter: filter,
	}
}

func (n *filterNode) Serialize(w io.Writer, indentationLevel int) {
	if n.filter == nil {
		n.Node.Serialize(w, indentationLevel)
		return
	}

	io.WriteString(w, fmt.Sprintf("%sfilter by %s\n", indent(indentationLevel), n.filter))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *filterNode) Optimize() {
	if n.filter != nil {
		n.filter = n.distributeFilter(n.filter.Fold())
	}

	n.Node.Optimize()
}

func (n *filterNode) distributeFilter(filter expressions.Expression) expressions.Expression {
	var conjunctions []expressions.Expression
	for _, expression := range filter.Conjunctions() {
		if !n.distributeExpression(expression) {
			conjunctions = append(conjunctions, expression)
		}
	}

	return combineConjunctions(conjunctions)
}

func (n *filterNode) distributeExpression(expression expressions.Expression) bool {
	return n.Node.PushDownFilter(expression)
}

func (n *filterNode) PushDownFilter(filter expressions.Expression) bool {
	if n.filter != nil {
		filter = expressions.NewAnd(n.filter, filter)
	}

	n.filter = filter
	return true
}

func (n *filterNode) Scan(visitor VisitorFunc) error {
	if n.filter == nil {
		return n.Node.Scan(visitor)
	}

	return n.Node.Scan(func(row shared.Row) (bool, error) {
		if ok, err := shared.EnsureBool(n.filter.ValueFrom(row)); err != nil {
			return false, err
		} else if !ok {
			return true, nil
		}

		return visitor(row)
	})
}
