package indexes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

func (i *hashIndex) Scanner(ctx queries.Context, opts HashIndexScanOptions) (tidScanner, error) {
	value, err := ctx.Evaluate(opts.expression, shared.Row{})
	if err != nil {
		return nil, err
	}

	items := i.entries[shared.Hash(value)]

	j := 0

	return tidScannerFunc(func() (int, error) {
		if j < len(items) {
			tid := items[j].tid
			j++
			return tid, nil
		}

		return 0, scan.ErrNoRows
	}), nil
}

func (i *hashIndex) extractTIDAndValueFromRow(row shared.Row) (int, any, error) {
	tid, err := row.TID()
	if err != nil {
		return 0, nil, err
	}

	value, err := i.expression.ValueFrom(expressions.EmptyContext, row)
	if err != nil {
		return 0, nil, err
	}

	return tid, value, nil
}
