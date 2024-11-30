package opt

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalInsertNode struct {
	LogicalNode
	table       impls.Table
	fields      []fields.Field
	columnNames []string
}

func NewInsert(node LogicalNode, table impls.Table, columnNames []string) (LogicalNode, error) {
	var fields []fields.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	return &logicalInsertNode{
		LogicalNode: node,
		table:       table,
		fields:      fields,
		columnNames: columnNames,
	}, nil
}

func (n *logicalInsertNode) Fields() []fields.Field                                              { return n.fields }
func (n *logicalInsertNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalInsertNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *logicalInsertNode) Filter() impls.Expression                                            { return nil }
func (n *logicalInsertNode) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalInsertNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalInsertNode) Build() nodes.Node {
	return nodes.NewInsert(n.LogicalNode.Build(), n.table, n.fields, n.columnNames)
}
