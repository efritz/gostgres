package filter

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/types"
)

func NewFilterScanner(ctx impls.ExecutionContext, scanner scan.RowScanner, filter impls.Expression) (scan.RowScanner, error) {
	ctx.Log("Building Filter scanner")

	if filter == nil {
		return scanner, nil
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Filter")

		for {
			row, err := scanner.Scan()
			if err != nil {
				return rows.Row{}, err
			}

			if ok, err := types.ValueAs[bool](queries.Evaluate(ctx, filter, row)); err != nil {
				return rows.Row{}, err
			} else if ok == nil || !*ok {
				continue
			}

			return row, nil
		}
	}), nil
}
