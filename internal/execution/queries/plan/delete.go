package plan

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalDeleteNode struct {
	LogicalNode
	table     impls.Table
	fields    []fields.Field
	aliasName string
}

func NewDelete(node LogicalNode, table impls.Table, aliasName string) (LogicalNode, error) {
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

func (n *logicalDeleteNode) Fields() []fields.Field                                              { return n.fields }
func (n *logicalDeleteNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalDeleteNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *logicalDeleteNode) Filter() impls.Expression                                            { return nil }
func (n *logicalDeleteNode) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalDeleteNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalDeleteNode) Build() nodes.Node {
	return nodes.NewDelete(n.LogicalNode.Build(), n.table, n.fields, n.aliasName)
}
