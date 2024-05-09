package sequence

import (
	"github.com/efritz/gostgres/internal/shared"
)

type Sequencespace struct {
	sequences map[string]*Sequence
}

func NewSequencespace() *Sequencespace {
	return &Sequencespace{
		sequences: map[string]*Sequence{},
	}
}

func (s *Sequencespace) GetSequence(name string) (*Sequence, bool) {
	table, ok := s.sequences[name]
	return table, ok
}

func (s *Sequencespace) CreateSequence(name string, typ shared.Type) error {
	_, err := s.CreateAndGetSequence(name, typ)
	return err
}

func (s *Sequencespace) CreateAndGetSequence(name string, typ shared.Type) (*Sequence, error) {
	sequence := NewSequence(name, typ)
	s.sequences[name] = sequence
	return sequence, nil
}
