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

type explain struct {
	n queries.Node
}

var _ queries.Node = &explain{}

func NewExplain(n queries.Node) *explain {
	return &explain{
		n: n,
	}
}

func (n *explain) Name() string {
	return "EXPLAIN"
}

var queryPlanField = fields.NewField("", "query plan", types.TypeText, fields.NonInternalField)

func (n *explain) Fields() []fields.Field {
	return []fields.Field{queryPlanField}
}

func (n *explain) Serialize(w serialization.IndentWriter)                              {}
func (n *explain) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *explain) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *explain) Optimize(ctx impls.OptimizationContext)                              { n.n.Optimize(ctx) }
func (n *explain) Filter() impls.Expression                                            { return nil }
func (n *explain) Ordering() impls.OrderExpression                                     { return nil }
func (n *explain) SupportsMarkRestore() bool                                           { return false }

func (n *explain) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Explain scanner")

	plan := serialization.SerializePlan(n.n)
	emitted := false

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Explain")

		if emitted {
			return rows.Row{}, scan.ErrNoRows
		}

		emitted = true
		return rows.NewRow(n.Fields(), []any{plan})
	}), nil
}
