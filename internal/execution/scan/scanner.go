package scan

import "github.com/efritz/gostgres/internal/shared/rows"

type Scanner interface {
	Scan() (rows.Row, error)
}

type MarkRestorer interface {
	Mark()
	Restore()
}

type ScannerFunc func() (rows.Row, error)

func (f ScannerFunc) Scan() (rows.Row, error) {
	return f()
}
