package nodes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type BaseIndex interface {
	Name() string
	Filter() expressions.Expression
	Ordering() OrderExpression
	Insert(row shared.Row) error
	Delete(row shared.Row) error
}

type ScanOptions interface {
	Condition() expressions.Expression
}

type IndexScanner[O ScanOptions] interface {
	Scanner(ctx ScanContext, opts O) (tidScanner, error)
}

type Index[O ScanOptions] interface {
	BaseIndex
	IndexScanner[O]
}

type tidScanner interface {
	Scan() (int, error)
}

type tidScannerFunc func() (int, error)

func (f tidScannerFunc) Scan() (int, error) {
	return f()
}

var EmptyTIDScanner = tidScannerFunc(func() (int, error) {
	return 0, ErrNoRows
})
