package functions

import (
	"github.com/efritz/gostgres/internal/catalog/sequence"
)

type FunctionContext interface {
	GetSequence(name string) (*sequence.Sequence, bool)
}
