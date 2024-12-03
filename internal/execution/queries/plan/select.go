package plan

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalSelectNode struct {
	LogicalNode
	projection *projection.Projection
	order      impls.OrderExpression
	limit      *int
	offset     *int
}

func NewSelect(
	node LogicalNode,
	projection *projection.Projection,
	order impls.OrderExpression,
	limit *int,
	offset *int,
) LogicalNode {
	return &logicalSelectNode{
		LogicalNode: node,
		projection:  projection,
		order:       order,
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

	// NUANCED!
	// We want to push the order down to the underlying node, but we don't want to
	// add an explicit ordering step here as it might just be wasted if we're also
	// ordering in a node above.

	if n.order != nil {
		// Blast away our own order
		//
		// We are nested in a parent sort and un-separated by an ordering boundary
		// (such as limit or offset). We'll ignore our old sort criteria and adopt
		// our parent since the ordering of rows at this point in the query should
		// not have an effect on the result.
		n.order = order
	}

	// Pro-actively push down the order into the underlying node so that it might
	// eventually hit access nodes and better inform choices of join strategies.
	n.LogicalNode.AddOrder(ctx, order)
}

func (n *logicalSelectNode) Optimize(ctx impls.OptimizationContext) {
	if n.projection != nil {
		n.projection.Optimize(ctx)
	}

	if n.order != nil {
		n.order = n.order.Fold()
		n.LogicalNode.AddOrder(ctx, n.order)
	}

	n.LogicalNode.Optimize(ctx)

	if expressions.SubsumesOrder(n.order, n.LogicalNode.Ordering()) {
		n.order = nil
	}
}

func (n *logicalSelectNode) Filter() impls.Expression {
	filter := n.LogicalNode.Filter()

	if n.projection != nil {
		filter = n.projection.ProjectExpression(filter)
	}

	return filter
}

func (n *logicalSelectNode) Ordering() impls.OrderExpression {
	ordering := n.order
	if ordering == nil {
		ordering = n.LogicalNode.Ordering()
	}
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

	if n.order != nil {
		node = nodes.NewOrder(node, n.order, n.LogicalNode.Fields())
	}

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
