package nodes

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type valuesNode struct {
	fields      []fields.Field
	expressions [][]impls.Expression
}

func NewValues(fields []fields.Field, expressions [][]impls.Expression) Node {
	return &valuesNode{
		fields:      fields,
		expressions: expressions,
	}
}

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
