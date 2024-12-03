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
	limit      *int
	offset     *int
}

func NewSelect(
	node LogicalNode,
	projection *projection.Projection,
	limit *int,
	offset *int,
) LogicalNode {
	return &logicalSelectNode{
		LogicalNode: node,
		projection:  projection,
		limit:       limit,
		offset:      offset,
	}
}

func (n *logicalSelectNode) Name() string {
	if n.projection != nil {
		return ""
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
	if n.limit != nil || n.offset != nil {
		return // boundary
	}

	if n.projection != nil {
		filter = n.projection.DeprojectExpression(filter)
	}

	n.LogicalNode.AddFilter(ctx, filter)
}

func (n *logicalSelectNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	if n.limit != nil || n.offset != nil {
		return // boundary
	}

	if n.projection != nil {
		order, _ = order.Map(func(expression impls.Expression) (impls.Expression, error) {
			return n.projection.DeprojectExpression(expression), nil
		})
	}

	n.LogicalNode.AddOrder(ctx, order)
}

func (n *logicalSelectNode) Optimize(ctx impls.OptimizationContext) {
	if n.projection != nil {
		n.projection.Optimize(ctx)
	}

	n.LogicalNode.Optimize(ctx)
}

func (n *logicalSelectNode) Filter() impls.Expression {
	filter := n.LogicalNode.Filter()

	if n.projection != nil {
		filter = n.projection.ProjectExpression(filter)
	}

	return filter
}

func (n *logicalSelectNode) Ordering() impls.OrderExpression {
	ordering := n.LogicalNode.Ordering()
	if ordering == nil {
		return nil
	}

	if n.projection != nil {
		ordering, _ = ordering.Map(func(expression impls.Expression) (impls.Expression, error) {
			return n.projection.ProjectExpression(expression), nil
		})
	}

	return ordering
}

func (n *logicalSelectNode) SupportsMarkRestore() bool {
	return false // TODO
}

func (n *logicalSelectNode) Build() nodes.Node {
	node := n.LogicalNode.Build()

	if n.offset != nil {
		node = nodes.NewOffset(node, *n.offset)
	}

	if n.limit != nil {
		node = nodes.NewLimit(node, *n.limit)
	}

	if n.projection != nil {
		node = nodes.NewProjection(node, n.projection)
	}

	return node
}
