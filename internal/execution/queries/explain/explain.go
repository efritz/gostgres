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

var queryPlanField = fields.ResolveDescriptor(fields.NewFieldDescriptor("", "query plan", types.TypeText, fields.NotInternal))

func (n *explain) Fields() []fields.ResolvedField {
	return []fields.ResolvedField{queryPlanField}
}

func (n *explain) Serialize(w serialization.IndentWriter) {}
func (n *explain) AddFilter(filter impls.Expression)      {}
func (n *explain) AddOrder(order impls.OrderExpression)   {}
func (n *explain) Optimize()                              { n.n.Optimize() }
func (n *explain) Filter() impls.Expression               { return nil }
func (n *explain) Ordering() impls.OrderExpression        { return nil }
func (n *explain) SupportsMarkRestore() bool              { return false }

func (n *explain) Scanner(ctx impls.Context) (scan.RowScanner, error) {
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
