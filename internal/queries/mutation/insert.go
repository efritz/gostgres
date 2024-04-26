package mutation

import (
	"fmt"
	"io"
	"slices"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/projection"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type insertNode struct {
	queries.Node
	table       *table.Table
	columnNames []string
	projector   *projection.Projector
}

var _ queries.Node = &insertNode{}

func NewInsert(node queries.Node, table *table.Table, name, alias string, columnNames []string, expressions []projection.ProjectionExpression) (queries.Node, error) {
	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(name, table.Fields(), alias)
		}
	}

	projector, err := projection.NewProjector(node.Name(), table.Fields(), expressions)
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
	return slices.Clone(n.projector.Fields())
}

func (n *insertNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sinsert returning (%s)\n", serialization.Indent(indentationLevel), n.projector))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *insertNode) Optimize() {
	n.projector.Optimize()
	n.Node.Optimize()
}

func (n *insertNode) AddFilter(filter expressions.Expression) {
}

func (n *insertNode) AddOrder(order expressions.OrderExpression) {
}

func (n *insertNode) Filter() expressions.Expression {
	return nil
}

func (n *insertNode) Ordering() expressions.OrderExpression {
	return nil
}

func (n *insertNode) SupportsMarkRestore() bool {
	return false
}

func (n *insertNode) Scanner(ctx queries.Context) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.ScannerFunc(func() (shared.Row, error) {
		row, err := scanner.Scan()
		if err != nil {
			return shared.Row{}, err
		}

		fields := make([]shared.Field, 0, len(n.table.Fields()))
		for _, field := range n.table.Fields() {
			if !field.Internal() {
				fields = append(fields, field)
			}
		}

		values, err := n.prepareValuesForRow(row, fields)
		if err != nil {
			return shared.Row{}, err
		}

		insertedRow, err := shared.NewRow(fields, values)
		if err != nil {
			return shared.Row{}, err
		}

		insertedRow, err = n.table.Insert(insertedRow)
		if err != nil {
			return shared.Row{}, err
		}

		return n.projector.ProjectRow(ctx, insertedRow)
	}), nil
}

func (n *insertNode) prepareValuesForRow(row shared.Row, fields []shared.Field) ([]any, error) {
	values := make([]any, 0, len(row.Values))
	for i, value := range row.Values {
		if !row.Fields[i].Internal() {
			values = append(values, value)
		}
	}

	if n.columnNames == nil {
		return values, nil
	}

	if len(values) != len(n.columnNames) {
		return nil, fmt.Errorf("number of columns does not match number of values")
	}

	valueMap := make(map[string]any, len(n.columnNames))
	for i, name := range n.columnNames {
		valueMap[name] = values[i]
	}

	reordered := make([]any, 0, len(fields))
	for _, field := range fields {
		value, ok := valueMap[field.Name()]
		if !ok {
			value, ok = field.Default()
			if !ok && !field.Type().Nullable {
				return nil, fmt.Errorf("no value supplied for %s", field.Name())
			}
		}

		reordered = append(reordered, value)
	}

	return reordered, nil
}
