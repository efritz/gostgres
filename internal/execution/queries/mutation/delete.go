package mutation

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type deleteNode struct {
	queries.Node
	table     impls.Table
	fields    []fields.Field
	aliasName string
}

var _ queries.Node = &deleteNode{}

func NewDelete(node queries.Node, table impls.Table, aliasName string) (queries.Node, error) {
	var fields []fields.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	return &deleteNode{
		Node:      node,
		table:     table,
		fields:    fields,
		aliasName: aliasName,
	}, nil
}

func (n *deleteNode) Fields() []fields.Field {
	return n.fields
}

func (n *deleteNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("delete from %s", n.table.Name())
	n.Node.Serialize(w.Indent())
}

func (n *deleteNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *deleteNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}

func (n *deleteNode) Optimize(ctx impls.OptimizationContext) {
	n.Node.Optimize(ctx)
}

func (n *deleteNode) Filter() impls.Expression        { return nil }
func (n *deleteNode) Ordering() impls.OrderExpression { return nil }
func (n *deleteNode) SupportsMarkRestore() bool       { return false }

func (n *deleteNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Delete scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Delete")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		relationName := n.table.Name()
		if n.aliasName != "" {
			relationName = n.aliasName
		}

		tidRow, err := row.IsolateTID(relationName)
		if err != nil {
			return rows.Row{}, err
		}

		deletedRow, ok, err := n.table.Delete(tidRow)
		if err != nil {
			return rows.Row{}, err
		}
		if !ok {
			return rows.Row{}, scan.ErrNoRows
		}

		return deletedRow, nil
	}), nil
}
