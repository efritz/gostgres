package plan

import (
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalSelectNode struct {
	LogicalNode
	projection *projection.Projection
}

func NewSelect(
	node LogicalNode,
	projection *projection.Projection,
) LogicalNode {
	return &logicalSelectNode{
		LogicalNode: node,
		projection:  projection,
	}
}

func (n *logicalSelectNode) Name() string {
	if n.projection != nil {
		return n.LogicalNode.Name()
	} else {
		return n.LogicalNode.Name()
	}
}

func (n *logicalSelectNode) Fields() []fields.Field {
	if n.projection != nil {
		return n.projection.Fields()
	} else {
		return n.LogicalNode.Fields()
	}
}

func (n *logicalSelectNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	if n.projection != nil {
		n.LogicalNode.AddFilter(ctx, n.projection.DeprojectExpression(filter))
	} else {
		n.LogicalNode.AddFilter(ctx, filter)
	}
}

func (n *logicalSelectNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	if n.projection != nil {
		mapped, _ := order.Map(func(expression impls.Expression) (impls.Expression, error) {
			return n.projection.DeprojectExpression(expression), nil
		})

		n.LogicalNode.AddOrder(ctx, mapped)
	} else {
		n.LogicalNode.AddOrder(ctx, order)
	}
}

func (n *logicalSelectNode) Optimize(ctx impls.OptimizationContext) {
	if n.projection != nil {
		n.projection.Optimize(ctx)
		n.LogicalNode.Optimize(ctx)
	} else {
		n.LogicalNode.Optimize(ctx)
	}
}

func (n *logicalSelectNode) Filter() impls.Expression {
	if n.projection != nil {
		return n.projection.ProjectExpression(n.LogicalNode.Filter())
	} else {
		return n.LogicalNode.Filter()
	}
}

func (n *logicalSelectNode) Ordering() impls.OrderExpression {
	if n.projection != nil {
		ordering := n.LogicalNode.Ordering()
		if ordering == nil {
			return nil
		}

		mapped, _ := ordering.Map(func(expression impls.Expression) (impls.Expression, error) {
			return n.projection.ProjectExpression(expression), nil
		})

		return mapped
	} else {
		return n.LogicalNode.Ordering()
	}
}

func (n *logicalSelectNode) SupportsMarkRestore() bool {
	if n.projection != nil {
		return false
	} else {
		return n.LogicalNode.SupportsMarkRestore()
	}
}

func (n *logicalSelectNode) Build() nodes.Node {
	if n.projection != nil {
		return nodes.NewProjection(n.LogicalNode.Build(), n.projection)
	} else {
		return n.LogicalNode.Build()
	}
}
