package scan

import "github.com/efritz/gostgres/internal/shared"

type Scanner interface {
	Scan() (shared.Row, error)
}

type MarkRestorer interface {
	Mark()
	Restore()
}

type ScannerFunc func() (shared.Row, error)

func (f ScannerFunc) Scan() (shared.Row, error) {
	return f()
}
