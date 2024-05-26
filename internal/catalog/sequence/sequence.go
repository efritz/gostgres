package sequence

import (
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type sequence struct {
	name  string
	typ   shared.Type // TODO - actually store as this type
	value int64
}

func NewSequence(name string, typ shared.Type) types.Sequence {
	return &sequence{
		name: name,
		typ:  typ,
	}
}

func (s *sequence) Name() string {
	return s.name
}

func (s *sequence) Next() (int64, error) {
	s.value++
	return s.value, nil
}

func (s *sequence) Set(value int64) error {
	s.value = value
	return nil
}

func (s *sequence) Value() int64 {
	return s.value
}
