package indexes

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/utils"
)

func (i *hashIndex) Scanner(ctx impls.Context, opts HashIndexScanOptions) (scan.TIDScanner, error) {
	ctx.Log("Building Hash Index scanner")

	value, err := queries.Evaluate(ctx, opts.expression, rows.Row{})
	if err != nil {
		return nil, err
	}

	items := i.entries[utils.Hash(value)]

	j := 0

	return scan.TIDScannerFunc(func() (int64, error) {
		ctx.Log("Scanning Hash Index")

		if j < len(items) {
			tid := items[j].tid
			j++
			return tid, nil
		}

		return 0, scan.ErrNoRows
	}), nil
}
