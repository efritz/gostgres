package explain

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/types"
)

type logicalExplain struct {
	n queries.LogicalNode
}

var _ queries.LogicalNode = &logicalExplain{}

func NewExplain(n queries.LogicalNode) *logicalExplain {
	return &logicalExplain{
		n: n,
	}
}

func (n *logicalExplain) Name() string {
	return "EXPLAIN"
}

var queryPlanField = fields.NewField("", "query plan", types.TypeText, fields.NonInternalField)

func (n *logicalExplain) Fields() []fields.Field {
	return []fields.Field{queryPlanField}
}

func (n *logicalExplain) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalExplain) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *logicalExplain) Optimize(ctx impls.OptimizationContext)                              { n.n.Optimize(ctx) }
func (n *logicalExplain) Filter() impls.Expression                                            { return nil }
func (n *logicalExplain) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalExplain) SupportsMarkRestore() bool                                           { return false }

func (n *logicalExplain) Build() queries.Node {
	return &explainNode{
		n:      n.n.Build(),
		fields: n.Fields(),
	}
}

//
//

type explainNode struct {
	n      queries.Node
	fields []fields.Field
}

var _ queries.Node = &explainNode{}

func (n *explainNode) Serialize(w serialization.IndentWriter) {
}

func (n *explainNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Explain scanner")

	plan := serialization.SerializePlan(n.n)
	emitted := false

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Explain")

		if emitted {
			return rows.Row{}, scan.ErrNoRows
		}

		emitted = true
		return rows.NewRow(n.fields, []any{plan})
	}), nil
}
