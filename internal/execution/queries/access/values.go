package access

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type valuesNode struct {
	fields      []fields.ResolvedField
	expressions [][]impls.Expression
}

var _ queries.Node = &valuesNode{}

func NewValues(fields []fields.ResolvedField, expressions [][]impls.Expression) queries.Node {
	return &valuesNode{
		fields:      fields,
		expressions: expressions,
	}
}

func (n *valuesNode) Name() string {
	return "values"
}

func (n *valuesNode) Fields() []fields.ResolvedField {
	return n.fields
}

func (n *valuesNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("values")
}

func (n *valuesNode) AddFilter(filter impls.Expression)    {}
func (n *valuesNode) AddOrder(order impls.OrderExpression) {}
func (n *valuesNode) Optimize()                            {}
func (n *valuesNode) Filter() impls.Expression             { return nil }
func (n *valuesNode) Ordering() impls.OrderExpression      { return nil }
func (n *valuesNode) SupportsMarkRestore() bool            { return false }

func (n *valuesNode) Scanner(ctx impls.Context) (scan.RowScanner, error) {
	ctx.Log("Building Values scanner")

	i := 0

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Values")

		if i >= len(n.expressions) {
			return rows.Row{}, scan.ErrNoRows
		}

		exprs := n.expressions[i]
		i++

		var values []any
		for _, expr := range exprs {
			value, err := queries.Evaluate(ctx, expr, rows.Row{})
			if err != nil {
				return rows.Row{}, err
			}

			values = append(values, value)
		}

		return rows.NewRow(n.fields, values)
	}), nil
}
