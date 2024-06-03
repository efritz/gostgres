package ast

import "github.com/efritz/gostgres/internal/shared/impls"

type BuildContext struct {
	Tables TableGetter
}

type TableGetter interface {
	Get(name string) (impls.Table, bool)
}
