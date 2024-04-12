package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type orderNode struct {
	Node
	order OrderExpression
}

var _ Node = &orderNode{}

func NewOrder(node Node, order OrderExpression) Node {
	return &orderNode{
		Node:  node,
		order: order,
	}
}

func (n *orderNode) Serialize(w io.Writer, indentationLevel int) {
	if n.order == nil {
		n.Node.Serialize(w, indentationLevel)
		return
	}

	io.WriteString(w, fmt.Sprintf("%sorder by %s\n", indent(indentationLevel), n.order))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *orderNode) Optimize() {
	if n.order != nil {
		n.order = n.order.Fold()
		n.Node.AddOrder(n.order)
	}

	n.Node.Optimize()

	if subsumesOrder(n.order, n.Node.Ordering()) {
		n.order = nil
	}
}

func subsumesOrder(a, b OrderExpression) bool {
	if a == nil || b == nil {
		return false
	}

	aExpressions := a.Expressions()
	bExpressions := b.Expressions()
	if len(bExpressions) < len(aExpressions) {
		return false
	}

	for i, expression := range aExpressions {
		if expression.Reverse != bExpressions[i].Reverse {
			return false
		}

		if !expression.Expression.Equal(bExpressions[i].Expression) {
			return false
		}
	}

	return true
}

func (n *orderNode) AddFilter(filter expressions.Expression) {
	n.Node.AddFilter(filter)
}

func (n *orderNode) AddOrder(order OrderExpression) {
	// We are nested in a parent sort and un-separated by an ordering boundary
	// (such as limit or offset). We'll ignore our old sort criteria and adopt
	// our parent since the ordering of rows at this point in the query should
	// not have an effect on the result.
	n.order = order
}

func (n *orderNode) Ordering() OrderExpression {
	if n.order == nil {
		return n.Node.Ordering()
	}

	return n.order
}

func (n *orderNode) SupportsMarkRestore() bool {
	return true
}

func (n *orderNode) Scanner() (Scanner, error) {
	scanner, err := n.Node.Scanner()
	if err != nil {
		return nil, err
	}

	// TODO - commented out to support mark-restore
	// if n.order == nil {
	// 	return scanner, nil
	// }

	return NewOrderScanner(scanner, n.Fields(), n.order)
}

type orderScanner struct {
	rows    shared.Rows
	indexes []int
	next    int
	mark    int
}

func NewOrderScanner(scanner Scanner, fields []shared.Field, order OrderExpression) (Scanner, error) {
	rows, err := shared.NewRows(fields)
	if err != nil {
		return nil, err
	}

	rows, err = ScanIntoRows(scanner, rows)
	if err != nil {
		return nil, err
	}

	indexes, err := findIndexIterationOrder(order, rows)
	if err != nil {
		return nil, err
	}

	return &orderScanner{
		rows:    rows,
		indexes: indexes,
		mark:    -1,
	}, nil
}

func (s *orderScanner) Scan() (shared.Row, error) {
	if s.next < len(s.indexes) {
		row := s.rows.Row(s.indexes[s.next])
		s.next++
		return row, nil
	}

	return shared.Row{}, ErrNoRows
}

func (s *orderScanner) Mark() {
	s.mark = s.next - 1
}

func (s *orderScanner) Restore() {
	if s.mark == -1 {
		panic("no mark to restore")
	}

	s.next = s.mark
}
