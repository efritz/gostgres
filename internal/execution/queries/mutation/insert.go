package mutation

import (
	"fmt"
	"slices"

	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type insertNode struct {
	queries.Node
	table       impls.Table
	columnNames []string
	projector   *projector.Projector
}

var _ queries.Node = &insertNode{}

func NewInsert(node queries.Node, table impls.Table, name, alias string, columnNames []string, expressions []projector.ProjectionExpression) (queries.Node, error) {
	var fields []fields.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	if alias != "" {
		for i, pe := range expressions {
			expressions[i] = pe.Dealias(name, fields, alias)
		}
	}

	projector, err := projector.NewProjector(node.Name(), fields, expressions)
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

func (n *insertNode) Fields() []fields.Field {
	return slices.Clone(n.projector.Fields())
}

func (n *insertNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("insert returning (%s)", n.projector)
	n.Node.Serialize(w.Indent())
}

func (n *insertNode) AddFilter(filter impls.Expression)    {}
func (n *insertNode) AddOrder(order impls.OrderExpression) {}

func (n *insertNode) Optimize() {
	n.projector.Optimize()
	n.Node.Optimize()
}

func (n *insertNode) Filter() impls.Expression        { return nil }
func (n *insertNode) Ordering() impls.OrderExpression { return nil }
func (n *insertNode) SupportsMarkRestore() bool       { return false }

func (n *insertNode) Scanner(ctx impls.Context) (scan.RowScanner, error) {
	ctx.Log("Building Insert scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var nonInternalFields []impls.TableField
	for _, field := range n.table.Fields() {
		if !field.Internal() {
			nonInternalFields = append(nonInternalFields, field)
		}
	}

	var fields []fields.Field
	for _, field := range nonInternalFields {
		fields = append(fields, field.Field)
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Insert")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		values, err := n.prepareValuesForRow(ctx, row, nonInternalFields)
		if err != nil {
			return rows.Row{}, err
		}

		insertedRow, err := rows.NewRow(fields, values)
		if err != nil {
			return rows.Row{}, err
		}

		insertedRow, err = n.table.Insert(ctx, insertedRow)
		if err != nil {
			return rows.Row{}, err
		}

		return n.projector.ProjectRow(ctx, insertedRow)
	}), nil
}

func (n *insertNode) prepareValuesForRow(ctx impls.Context, row rows.Row, fields []impls.TableField) ([]any, error) {
	values := make([]any, 0, len(row.Values))
	for i, value := range row.Values {
		if !row.Fields[i].Internal() {
			values = append(values, value)
		}
	}

	if n.columnNames == nil {
		// TODO - still need to check nullability?
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
			defaultValue, err := field.Default(ctx)
			if err != nil {
				return nil, err
			}

			value = defaultValue
		}

		if value == nil && !field.Nullable() {
			return nil, fmt.Errorf("%s is not a nullable column", field.Name())
		}

		reordered = append(reordered, value)
	}

	return reordered, nil
}
