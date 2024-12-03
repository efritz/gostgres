package nodes

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type insertNode struct {
	Node
	table       impls.Table
	columnNames []string
}

func NewInsert(node Node, table impls.Table, columnNames []string) Node {
	return &insertNode{
		Node:        node,
		table:       table,
		columnNames: columnNames,
	}
}

func (n *insertNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("insert into %s", n.table.Name())
	n.Node.Serialize(w.Indent())
}

func (n *insertNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
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

		return insertedRow, nil
	}), nil
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
