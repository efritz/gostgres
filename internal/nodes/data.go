package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

// TODO - rename to table scan/access (depending on if we use indexes too)
type dataNode struct {
	table  *Table
	filter expressions.Expression
	order  OrderExpression
}

var _ Node = &dataNode{}

func NewData(table *Table) Node {
	return &dataNode{
		table: table,
	}
}

func (n *dataNode) Name() string {
	return n.table.name
}

func (n *dataNode) Fields() []shared.Field {
	return updateRelationName(n.table.Fields(), n.table.name)
}

func (n *dataNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%saccess of %s\n", indent(indentationLevel), n.table.name))
	if n.filter != nil {
		io.WriteString(w, fmt.Sprintf("%sfilter: %s\n", indent(indentationLevel+1), n.filter))
	}
	if n.order != nil {
		io.WriteString(w, fmt.Sprintf("%sorder: %s\n", indent(indentationLevel+1), n.order))
	}
}

func (n *dataNode) Optimize() {
	if n.filter != nil {
		n.filter = n.filter.Fold()
	}

	if n.order != nil {
		n.order = n.order.Fold()
	}
}

func (n *dataNode) AddFilter(filter expressions.Expression) {
	n.filter = unionFilters(n.filter, filter)
}

func (n *dataNode) AddOrder(order OrderExpression) {
	if n.order != nil {
		panic("unreachable")
	}

	n.order = order
}

func (n *dataNode) Filter() expressions.Expression {
	return n.filter
}

func (n *dataNode) Ordering() OrderExpression {
	return n.order
}

func (n *dataNode) SupportsMarkRestore() bool {
	return false
}

func (n *dataNode) Scanner() (Scanner, error) {
	indexes, err := findIndexIterationOrder(n.order, n.table.rows)
	if err != nil {
		return nil, err
	}

	// Keep a copy so we don't modify it while iterating a subquery of a delete
	// statement. This will probably be able to remove once we have MVCC semantics
	// on each tuple, and may let us share the following logic with filter and
	// order nodes.

	rows := make([]shared.Row, 0, len(indexes))
	for _, i := range indexes {
		rows = append(rows, n.table.rows.Row(i))
	}

	i := 0

	return ScannerFunc(func() (shared.Row, error) {
		for i < len(indexes) {
			row := rows[i]
			i++

			if n.filter != nil {
				if ok, err := shared.EnsureBool(n.filter.ValueFrom(row)); err != nil {
					return shared.Row{}, err
				} else if !ok {
					continue
				}
			}

			return row, nil
		}

		return shared.Row{}, ErrNoRows
	}), nil
}
