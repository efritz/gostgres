package indexes

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

func (i *hashIndex) Scanner(ctx types.Context, opts HashIndexScanOptions) (tidScanner, error) {
	value, err := queries.Evaluate(ctx, opts.expression, shared.Row{})
	if err != nil {
		return nil, err
	}

	items := i.entries[shared.Hash(value)]

	j := 0

	return tidScannerFunc(func() (int64, error) {
		if j < len(items) {
			tid := items[j].tid
			j++
			return tid, nil
		}

		return 0, scan.ErrNoRows
	}), nil
}
