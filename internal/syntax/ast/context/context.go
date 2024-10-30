package context

import "github.com/efritz/gostgres/internal/shared/impls"

type ResolveContext struct {
	Tables TableGetter
}

type TableGetter interface {
	Get(name string) (impls.Table, bool)
}
