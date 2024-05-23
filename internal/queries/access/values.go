package access

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
)

type valuesNode struct {
	fields      []shared.Field
	expressions [][]expressions.Expression
}

var _ queries.Node = &valuesNode{}

func NewValues(fields []shared.Field, expressions [][]expressions.Expression) queries.Node {
	return &valuesNode{
		fields:      fields,
		expressions: expressions,
	}
}

func (n *valuesNode) Name() string {
	return "values"
}

func (n *valuesNode) Fields() []shared.Field {
	return n.fields
}

func (n *valuesNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("values")
}

func (n *valuesNode) AddFilter(filter expressions.Expression)    {}
func (n *valuesNode) AddOrder(order expressions.OrderExpression) {}
func (n *valuesNode) Optimize()                                  {}
func (n *valuesNode) Filter() expressions.Expression             { return nil }
func (n *valuesNode) Ordering() expressions.OrderExpression      { return nil }
func (n *valuesNode) SupportsMarkRestore() bool                  { return false }

func (n *valuesNode) Scanner(ctx queries.Context) (scan.Scanner, error) {
	i := 0

	return scan.ScannerFunc(func() (shared.Row, error) {
		if i >= len(n.expressions) {
			return shared.Row{}, scan.ErrNoRows
		}

		exprs := n.expressions[i]
		i++

		var values []any
		for _, expr := range exprs {
			value, err := queries.Evaluate(ctx, expr, shared.Row{})
			if err != nil {
				return shared.Row{}, err
			}

			values = append(values, value)
		}

		return shared.NewRow(n.fields, values)
	}), nil
}
