package mutation

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	projectionHelpers "github.com/efritz/gostgres/internal/execution/queries/nodes/projection"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type insertNode struct {
	nodes.Node
	table       impls.Table
	columnNames []string
	projection  *projection.Projection
}

func NewInsert(node nodes.Node, table impls.Table, columnNames []string, projection *projection.Projection) nodes.Node {
	return &insertNode{
		Node:        node,
		table:       table,
		columnNames: columnNames,
		projection:  projection,
	}
}

func (n *insertNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("insert into %s", n.table.Name())
	n.Node.Serialize(w.Indent())

	if n.projection != nil {
		w.WritefLine("returning %s", n.projection)
		n.Node.Serialize(w.Indent())
	}
}

func (n *insertNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Insert scanner")

	insertedRows, err := n.insertRows(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Insert")

		if len(insertedRows) != 0 {
			return rows.Row{}, scan.ErrNoRows
		}

		row := insertedRows[0]
		insertedRows = insertedRows[1:]
		return projectionHelpers.Project(ctx, row, n.projection)
	}), nil
}

func (n *insertNode) insertRows(ctx impls.ExecutionContext) ([]rows.Row, error) {
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

	var insertedRows []rows.Row
	for {
		row, err := scanner.Scan()
		if err != nil {
			if err == scan.ErrNoRows {
				break
			}

			return nil, err
		}

		values, err := n.prepareValuesForRow(ctx, row, nonInternalFields)
		if err != nil {
			return nil, err
		}

		insertedRow, err := rows.NewRow(fields, values)
		if err != nil {
			return nil, err
		}

		insertedRow, err = n.table.Insert(ctx, insertedRow)
		if err != nil {
			return nil, err
		}

		if n.projection != nil {
			insertedRows = append(insertedRows, insertedRow)
		}
	}

	return insertedRows, nil
}

func (n *insertNode) prepareValuesForRow(ctx impls.ExecutionContext, row rows.Row, fields []impls.TableField) ([]any, error) {
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
