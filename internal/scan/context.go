package scan

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type ScanContext struct {
	OuterRow shared.Row
}

func (ctx ScanContext) Evaluate(expr expressions.Expression, row shared.Row) (any, error) {
	return expr.ValueFrom(shared.CombineRows(row, ctx.OuterRow))
}
