package context

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type SymbolTable struct {
	Scopes []*Scope
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{}
}

func (st *SymbolTable) PushScope() {
	st.Scopes = append(st.Scopes, NewScope())
}

func (st *SymbolTable) PopScope() {
	st.Scopes = st.Scopes[:len(st.Scopes)-1]
}

func (st *SymbolTable) CurrentScope() *Scope {
	return st.Scopes[len(st.Scopes)-1]
}

func (st *SymbolTable) AddRelation(name, alias string, expressions []impls.Expression, fields []fields.Field) (Relation, error) {
	return st.CurrentScope().AddRelation(name, alias, expressions, fields)
}

type Scope struct {
	Symbols map[string]Relation
}

type Relation struct {
	Name        string
	UniqueName  string
	Expressions []impls.Expression
	Fields      []fields.Field // TODO - also track expressions?
}

func NewScope() *Scope {
	return &Scope{
		Symbols: map[string]Relation{},
	}
}

var i = 0

func (s *Scope) AddRelation(name, alias string, expressions []impls.Expression, fields []fields.Field) (Relation, error) {
	if _, ok := s.Symbols[alias]; ok {
		return Relation{}, fmt.Errorf("relation %q already defined", alias)
	}

	i++
	uniqueName := fmt.Sprintf("%s_%d", name, i)

	relation := Relation{Name: alias, UniqueName: uniqueName, Expressions: expressions, Fields: fields}
	s.Symbols[alias] = relation
	return relation, nil
}
