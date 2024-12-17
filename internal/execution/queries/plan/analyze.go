package plan

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalAnalyze struct {
	tables []impls.Table
}

func NewAnalyze(tables []impls.Table) LogicalNode {
	return &logicalAnalyze{
		tables: tables,
	}
}

func (n *logicalAnalyze) Name() string                                                        { return "ANALYZE" }
func (n *logicalAnalyze) Fields() []fields.Field                                              { return []fields.Field{} }
func (n *logicalAnalyze) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {} // top-level
func (n *logicalAnalyze) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {} // top-level
func (n *logicalAnalyze) Optimize(ctx impls.OptimizationContext)                              {}
func (n *logicalAnalyze) EstimateCost() Cost                                                  { return Cost{} }
func (n *logicalAnalyze) Filter() impls.Expression                                            { return nil }
func (n *logicalAnalyze) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalAnalyze) SupportsMarkRestore() bool                                           { return false }

func (n *logicalAnalyze) Build() nodes.Node {
	return nodes.NewAnalyze(n.tables)
}
