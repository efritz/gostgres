package context

import "github.com/efritz/gostgres/internal/shared/impls"

type ResolverContext = struct {
	Tables      TableGetter
	SymbolTable *SymbolTable
}

type TableGetter interface {
	Get(name string) (impls.Table, bool)
}

func NewResolverContext(tables TableGetter) *ResolverContext {
	return &ResolverContext{
		Tables:      tables,
		SymbolTable: NewSymbolTable(),
	}
}
