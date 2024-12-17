package join

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
)

// TODO - gomockgen

type mockLogicalNode struct{ name string }

func newMockLogicalNode(name string) *mockLogicalNode {
	return &mockLogicalNode{name: name}
}

func (m *mockLogicalNode) Name() string {
	return m.name
}

func (m *mockLogicalNode) Fields() []fields.Field {
	return []fields.Field{
		fields.NewField(m.name, "id", types.TypeBigInteger, fields.NonInternalField),
		fields.NewField(m.name, "foo", types.TypeBigInteger, fields.NonInternalField),
		fields.NewField(m.name, "bar", types.TypeBigInteger, fields.NonInternalField),
		fields.NewField(m.name, "baz", types.TypeBigInteger, fields.NonInternalField),
	}
}

func (*mockLogicalNode) AddFilter(impls.OptimizationContext, impls.Expression)     {}
func (*mockLogicalNode) AddOrder(impls.OptimizationContext, impls.OrderExpression) {}
func (*mockLogicalNode) Optimize(impls.OptimizationContext)                        {}
func (*mockLogicalNode) EstimateCost() plan.Cost                                   { return plan.Cost{} }
func (*mockLogicalNode) Filter() impls.Expression                                  { return nil }
func (*mockLogicalNode) Ordering() impls.OrderExpression                           { return nil }
func (*mockLogicalNode) SupportsMarkRestore() bool                                 { return false }
func (*mockLogicalNode) Build() nodes.Node                                         { return nil }
