package ast

import "github.com/efritz/gostgres/internal/shared/impls"

type ResolveContext struct {
	Tables TableGetter
}

type TableGetter interface {
	Get(name string) (impls.Table, bool)
}

type Resolver interface {
	Resolve(ctx ResolveContext) error
}
