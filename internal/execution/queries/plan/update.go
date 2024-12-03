package plan

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalUpdateNode struct {
	LogicalNode
	table          impls.Table
	fields         []fields.Field
	aliasName      string
	setExpressions []nodes.SetExpression
	filter         impls.Expression
}

func NewUpdate(node LogicalNode, table impls.Table, aliasName string, setExpressions []nodes.SetExpression, filter impls.Expression) (LogicalNode, error) {
	var fields []fields.Field
	for _, field := range table.Fields() {
		fields = append(fields, field.Field)
	}

	return &logicalUpdateNode{
		LogicalNode:    node,
		table:          table,
		fields:         fields,
		aliasName:      aliasName,
		setExpressions: setExpressions,
		filter:         filter,
	}, nil
}

func (n *logicalUpdateNode) Fields() []fields.Field                                              { return n.fields }
func (n *logicalUpdateNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalUpdateNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *logicalUpdateNode) Filter() impls.Expression                                            { return nil }
func (n *logicalUpdateNode) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalUpdateNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalUpdateNode) Optimize(ctx impls.OptimizationContext) {
	n.LogicalNode.Optimize(ctx)
	n.filter = n.filter.Fold()
}

func (n *logicalUpdateNode) Build() nodes.Node {
	node := n.LogicalNode.Build()
	if n.filter != nil {
		node = nodes.NewFilter(node, n.filter)
	}

	return nodes.NewUpdate(node, n.table, n.fields, n.aliasName, n.setExpressions)
}
