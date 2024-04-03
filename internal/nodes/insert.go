package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type insertNode struct {
	Node
	table       *Table
	columnNames []string
	projector   *projector
}

var _ Node = &insertNode{}

func NewInsert(node Node, table *Table, name, alias string, columnNames []string, expressions []ProjectionExpression) (Node, error) {
	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(name, table.Fields(), alias)
		}
	}

	projector, err := newProjector(node.Name(), table.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &insertNode{
		Node:        node,
		table:       table,
		columnNames: columnNames,
		projector:   projector,
	}, nil
}

func (n *insertNode) Fields() []shared.Field {
	return copyFields(n.projector.projectedFields)
}

func (n *insertNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sinsert returning (%s)\n", indent(indentationLevel), n.projector))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *insertNode) Optimize() {
	n.projector.optimize()
	n.Node.Optimize()
}

func (n *insertNode) AddFilter(filter expressions.Expression) {
}

func (n *insertNode) AddOrder(order OrderExpression) {
}

func (n *insertNode) Filter() expressions.Expression {
	return nil
}

func (n *insertNode) Ordering() OrderExpression {
	return nil
}

func (n *insertNode) Scanner() (Scanner, error) {
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

			fields := make([]shared.Field, 0, len(n.table.Fields()))
			for _, field := range n.table.Fields() {
				if !field.Internal {
					fields = append(fields, field)
				}
			}

			insertedRow, err := shared.NewRow(fields, n.prepareValuesForRow(row, fields))
			if err != nil {
				return shared.Row{}, err
			}

			insertedRow, err = n.table.Insert(insertedRow)
			if err != nil {
				return shared.Row{}, err
			}

			// TODO - necessary?
			if len(n.projector.aliases) == 0 {
				return shared.Row{}, nil
			}

			return n.projector.projectRow(insertedRow)
		}
	}), nil
}

func (n *insertNode) prepareValuesForRow(row shared.Row, fields []shared.Field) []interface{} {
	values := make([]interface{}, 0, len(row.Values))
	for i, value := range row.Values {
		if !row.Fields[i].Internal {
			values = append(values, value)
		}
	}

	if n.columnNames == nil {
		return values
	}

	return reorderValues(n.columnNames, values, fields)
}

func reorderValues(columnNames []string, values []interface{}, fields []shared.Field) []interface{} {
	valueMap := make(map[string]interface{}, len(columnNames))
	for i, name := range columnNames {
		valueMap[name] = values[i]
	}

	reordered := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		reordered = append(reordered, valueMap[field.Name])
	}

	return reordered
}
