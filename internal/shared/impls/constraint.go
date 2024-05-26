package impls

import "github.com/efritz/gostgres/internal/shared/rows"

type Constraint interface {
	Name() string
	Check(ctx Context, row rows.Row) error
}
