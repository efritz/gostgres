package functions

import "github.com/efritz/gostgres/internal/sequence"

type FunctionContext interface {
	GetFunction(name string) (Function, bool)
	GetSequence(name string) (*sequence.Sequence, bool)
}