package impls

import "github.com/efritz/gostgres/internal/shared/rows"

type Constraint interface {
	Name() string
	Check(ctx ExecutionContext, row rows.Row) error
}
