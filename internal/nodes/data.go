package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type dataNode struct {
	table *Table

	filter expressions.Expression
	order  OrderExpression
	// TODO - track fields for index only scans as well
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
	scanType := "seq"
	if n.filter != nil {
		scanType = "index"
	}

	io.WriteString(w, fmt.Sprintf("%s%s scan over %s\n", indent(indentationLevel), scanType, n.table.name))
	if n.filter != nil {
		io.WriteString(w, fmt.Sprintf("%sfilter: %s\n", indent(indentationLevel+1), n.filter))
	}
	if n.order != nil {
		// TODO - not yet populated
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

func (n *dataNode) PushDownFilter(filter expressions.Expression) bool {
	if n.filter != nil {
		filter = expressions.NewAnd(n.filter, filter)
	}

	n.filter = filter
	return true
}

func (n *dataNode) Scan(visitor VisitorFunc) error {
	// TODO - scan order needs to be aliased?
	indexes, err := findIndexIterationOrder(n.order, n.table.rows)
	if err != nil {
		return err
	}

	// Keep a copy so we don't modify it while iterating a subquery
	// of a delete statement. This will probably be able to remove
	// once we have MVCC semantics on each tuple.
	rows := make([]shared.Row, 0, len(indexes))
	for _, i := range indexes {
		rows = append(rows, n.table.rows.Row(i))
	}

	for _, row := range rows {
		if n.filter != nil {
			if ok, err := shared.EnsureBool(n.filter.ValueFrom(row)); err != nil {
				return err
			} else if !ok {
				continue
			}
		}

		if ok, err := visitor(row); err != nil {
			return err
		} else if !ok {
			break
		}
	}

	return nil
}
