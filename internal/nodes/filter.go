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
		n.filter = n.filter.Fold()
		n.Node.AddFilter(n.filter)
	}

	n.Node.Optimize()
	n.filter = filterDifference(n.filter, n.Node.Filter())
}

func (n *filterNode) AddFilter(filter expressions.Expression) {
	n.filter = unionFilters(n.filter, filter)
}

func (n *filterNode) AddOrder(order OrderExpression) {
	n.Node.AddOrder(order)
}

func (n *filterNode) Filter() expressions.Expression {
	return unionFilters(n.filter, n.Node.Filter())
}

func (n *filterNode) Ordering() OrderExpression {
	return n.Node.Ordering()
}

func (n *filterNode) Scanner() (Scanner, error) {
	scanner, err := n.Node.Scanner()
	if err != nil {
		return nil, err
	}

	if n.filter == nil {
		return scanner, nil
	}

	return ScannerFunc(func() (shared.Row, error) {
		for {
			row, err := scanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			if ok, err := shared.EnsureBool(n.filter.ValueFrom(row)); err != nil {
				return shared.Row{}, err
			} else if !ok {
				continue
			}

			return row, nil
		}
	}), nil
}
