package access

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalValuesNode struct {
	fields      []fields.Field
	expressions [][]impls.Expression
}

var _ queries.LogicalNode = &logicalValuesNode{}

func NewValues(fields []fields.Field, expressions [][]impls.Expression) queries.LogicalNode {
	return &logicalValuesNode{
		fields:      fields,
		expressions: expressions,
	}
}

func (n *logicalValuesNode) Name() string {
	return "values"
}

func (n *logicalValuesNode) Fields() []fields.Field {
	return n.fields
}

func (n *logicalValuesNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalValuesNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *logicalValuesNode) Optimize(ctx impls.OptimizationContext)                              {}
func (n *logicalValuesNode) Filter() impls.Expression                                            { return nil }
func (n *logicalValuesNode) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalValuesNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalValuesNode) Build() queries.Node {
	return &valuesNode{
		fields:      n.fields,
		expressions: n.expressions,
	}
}

//
//

type valuesNode struct {
	fields      []fields.Field
	expressions [][]impls.Expression
}

var _ queries.Node = &valuesNode{}

func (n *valuesNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("values")
}

func (n *valuesNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
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
