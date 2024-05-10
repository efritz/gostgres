package sequence

import (
	"github.com/efritz/gostgres/internal/shared"
)

type Sequence struct {
	name  string
	typ   shared.Type // TODO - actually store as this type
	value int64
}

func NewSequence(name string, typ shared.Type) *Sequence {
	return &Sequence{
		name: name,
		typ:  typ,
	}
}

func (s *Sequence) Name() string {
	return s.name
}

func (s *Sequence) Next() (int64, error) {
	s.value++
	return s.value, nil
}

func (s *Sequence) Set(value int64) error {
	s.value = value
	return nil
}

func (s *Sequence) Value() int64 {
	return s.value
}
