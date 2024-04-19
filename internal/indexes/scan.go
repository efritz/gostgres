package indexes

import "github.com/efritz/gostgres/internal/scan"

type tidScanner interface {
	Scan() (int, error)
}

type tidScannerFunc func() (int, error)

func (f tidScannerFunc) Scan() (int, error) {
	return f()
}

var EmptyTIDScanner = tidScannerFunc(func() (int, error) {
	return 0, scan.ErrNoRows
})
