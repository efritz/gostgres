package plan

import (
	"github.com/efritz/gostgres/internal/catalog/table/indexes"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/access"
	"github.com/efritz/gostgres/internal/execution/queries/plan/cost"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalAccessNode struct {
	table    impls.Table
	filter   impls.Expression
	order    impls.OrderExpression
	strategy logicalAccessStrategy
}

func NewAccess(table impls.Table) LogicalNode {
	return &logicalAccessNode{
		table: table,
	}
}

func (n *logicalAccessNode) Name() string {
	return n.table.Name()
}

func (n *logicalAccessNode) Fields() []fields.Field {
	var fields []fields.Field
	for _, field := range n.table.Fields() {
		// TODO - should never not be the case?
		field := field.WithRelationName(n.table.Name())
		fields = append(fields, field.Field)
	}

	return fields
}

func (n *logicalAccessNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filterExpression)
}

func (n *logicalAccessNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	n.order = order
}

func (n *logicalAccessNode) Optimize(ctx impls.OptimizationContext) {
	if n.filter != nil {
		n.filter = n.filter.Fold()
	}

	if n.order != nil {
		n.order = n.order.Fold()
	}

	n.strategy = selectAccessStrategy(n.table, n.filter, n.order)
	n.filter = expressions.FilterDifference(n.filter, n.strategy.Filter())
	n.order = nil
}

func (n *logicalAccessNode) EstimateCost() impls.NodeCost {
	return cost.ApplyFilterToCost(n.strategy.EstimateCost(), n.filter)
}

func (n *logicalAccessNode) Filter() impls.Expression {
	if filterExpression := n.strategy.Filter(); filterExpression != nil {
		return expressions.UnionFilters(n.filter, filterExpression)
	}

	return n.filter
}

func (n *logicalAccessNode) Ordering() impls.OrderExpression {
	return n.strategy.Ordering()
}

func (n *logicalAccessNode) SupportsMarkRestore() bool {
	return false
}

func (n *logicalAccessNode) Build() nodes.Node {
	node := nodes.NewAccess(n.strategy.Build())

	if n.filter != nil {
		node = nodes.NewFilter(node, n.filter)
	}

	return node
}

//
//

func selectAccessStrategy(
	table impls.Table,
	filterExpression impls.Expression,
	orderExpression impls.OrderExpression,
) logicalAccessStrategy {
	var candidates []logicalAccessStrategy
	for _, index := range table.Indexes() {
		if index, opts, ok := indexes.CanSelectHashIndex(index, filterExpression); ok {
			candidates = append(candidates, &logicalIndexAccessStrategy[indexes.HashIndexScanOptions]{table: table, index: index, opts: opts})
		}

		if index, opts, ok := indexes.CanSelectBtreeIndex(index, filterExpression, orderExpression); ok {
			candidates = append(candidates, &logicalIndexAccessStrategy[indexes.BtreeIndexScanOptions]{table: table, index: index, opts: opts})
		}
	}

	// TODO - use actual costs here
	maxScore := 0
	var bestStrategy logicalAccessStrategy = &logicalTableAccessStrategy{table: table}

	for _, index := range candidates {
		indexCost := 0
		if expressions.SubsumesOrder(orderExpression, index.Ordering()) {
			indexCost += 100
		}

		if filterExpression != nil {
			oldLen := len(expressions.Conjunctions(filterExpression))
			newLen := 0
			if remainingFilter := expressions.FilterDifference(filterExpression, index.Filter()); remainingFilter != nil {
				newLen = len(expressions.Conjunctions(remainingFilter))
			}

			indexCost += (oldLen - newLen) * 10
		}

		if indexCost > maxScore {
			bestStrategy = index
			maxScore = indexCost
		}
	}

	return bestStrategy
}

//
//

type logicalAccessStrategy interface {
	Filter() impls.Expression
	Ordering() impls.OrderExpression
	EstimateCost() impls.NodeCost
	Build() nodes.AccessStrategy
}

//
//

type logicalTableAccessStrategy struct {
	table impls.Table
}

func (s *logicalTableAccessStrategy) Filter() impls.Expression {
	return nil
}

func (s *logicalTableAccessStrategy) Ordering() impls.OrderExpression {
	return nil
}

var tableAccessCostPerRow = impls.ResourceCost{CPU: 0.01, IO: 0.1}

func (s *logicalTableAccessStrategy) EstimateCost() impls.NodeCost {
	stats := s.table.Statistics()

	return impls.NodeCost{
		VariableCost: tableAccessCostPerRow.ScaleUniform(float64(stats.RowCount)),
		Statistics:   stats,
	}
}

func (s *logicalTableAccessStrategy) Build() nodes.AccessStrategy {
	return access.NewTableAccessStrategy(s.table)
}

//
//

type logicalIndexAccessStrategy[O impls.ScanOptions] struct {
	table impls.Table
	index impls.Index[O]
	opts  O
}

func (s *logicalIndexAccessStrategy[O]) Ordering() impls.OrderExpression {
	return s.index.Ordering(s.opts)
}

func (s *logicalIndexAccessStrategy[O]) Filter() impls.Expression {
	filterExpression := s.index.Filter()
	condition := s.index.Condition(s.opts)

	return expressions.UnionFilters(append(expressions.Conjunctions(filterExpression), expressions.Conjunctions(condition)...)...)
}

var indexAccessCostPerRow = impls.ResourceCost{CPU: 0.01, IO: 0.1}

func (s *logicalIndexAccessStrategy[O]) EstimateCost() impls.NodeCost {
	tableStats := s.table.Statistics()
	indexStats := s.index.Statistics()

	// TODO - remove this use
	selectivity := cost.EstimateFilterSelectivity(s.index.Condition(s.opts), impls.RelationStatistics{
		RowCount:         indexStats.RowCount,
		ColumnStatistics: tableStats.ColumnStatistics,
	})
	estimatedRows := float64(indexStats.RowCount) * selectivity

	return impls.NodeCost{
		VariableCost: indexAccessCostPerRow.ScaleUniform(estimatedRows),
		Statistics: impls.RelationStatistics{
			RowCount:         int(estimatedRows),
			ColumnStatistics: tableStats.ColumnStatistics, // TODO - update based on filter, condition
		},
	}
}

func (s *logicalIndexAccessStrategy[O]) Build() nodes.AccessStrategy {
	return access.NewIndexAccessStrategy(s.table, s.index, s.opts, s.Filter())
}
