package scan

import "github.com/efritz/gostgres/internal/shared/rows"

type RowScanner interface {
	Scan() (rows.Row, error)
}

type RowScannerFunc func() (rows.Row, error)

func (f RowScannerFunc) Scan() (rows.Row, error) {
	return f()
}
