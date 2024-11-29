package mutation

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalDeleteNode struct {
	queries.LogicalNode
	table     impls.Table
	fields    []fields.Field
	aliasName string
}

var _ queries.LogicalNode = &logicalDeleteNode{}

func NewDelete(node queries.LogicalNode, table impls.Table, aliasName string) (queries.LogicalNode, error) {
	var fields []fields.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	return &logicalDeleteNode{
		LogicalNode: node,
		table:       table,
		fields:      fields,
		aliasName:   aliasName,
	}, nil
}

func (n *logicalDeleteNode) Fields() []fields.Field {
	return n.fields
}

func (n *logicalDeleteNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalDeleteNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}

func (n *logicalDeleteNode) Optimize(ctx impls.OptimizationContext) {
	n.LogicalNode.Optimize(ctx)
}

func (n *logicalDeleteNode) Filter() impls.Expression        { return nil }
func (n *logicalDeleteNode) Ordering() impls.OrderExpression { return nil }
func (n *logicalDeleteNode) SupportsMarkRestore() bool       { return false }

func (n *logicalDeleteNode) Build() queries.Node {
	return &deleteNode{
		Node:      n.LogicalNode.Build(),
		table:     n.table,
		fields:    n.fields,
		aliasName: n.aliasName,
	}
}

//
//

type deleteNode struct {
	queries.Node
	table     impls.Table
	fields    []fields.Field
	aliasName string
}

var _ queries.Node = &deleteNode{}

func (n *deleteNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("delete from %s", n.table.Name())
	n.Node.Serialize(w.Indent())
}

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
