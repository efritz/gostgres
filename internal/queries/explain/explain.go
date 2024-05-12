package explain

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
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

func (n *explain) Fields() []shared.Field {
	return []shared.Field{
		shared.NewField("", "query plan", shared.TypeText),
	}
}

func (n *explain) Serialize(w serialization.IndentWriter)     {}
func (n *explain) Optimize()                                  { n.n.Optimize() }
func (n *explain) AddFilter(filter expressions.Expression)    {}
func (n *explain) AddOrder(order expressions.OrderExpression) {}
func (n *explain) Filter() expressions.Expression             { return nil }
func (n *explain) Ordering() expressions.OrderExpression      { return nil }
func (n *explain) SupportsMarkRestore() bool                  { return false }

func (n *explain) Scanner(ctx queries.Context) (scan.Scanner, error) {
	plan := serialization.SerializePlan(n.n)
	emitted := false

	return scan.ScannerFunc(func() (shared.Row, error) {
		if emitted {
			return shared.Row{}, scan.ErrNoRows
		}

		emitted = true
		return shared.NewRow(n.Fields(), []any{plan})
	}), nil
}
