package filter

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/shared"
)

func NewFilterScanner(ctx queries.Context, scanner scan.Scanner, filter expressions.Expression) (scan.Scanner, error) {
	if filter == nil {
		return scanner, nil
	}

	return scan.ScannerFunc(func() (shared.Row, error) {
		for {
			row, err := scanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			if ok, err := shared.ValueAs[bool](queries.Evaluate(ctx, filter, row)); err != nil {
				return shared.Row{}, err
			} else if ok == nil || !*ok {
				continue
			}

			return row, nil
		}
	}), nil
}
