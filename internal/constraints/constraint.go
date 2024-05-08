package constraints

import "github.com/efritz/gostgres/internal/shared"

type Constraint interface {
	Name() string
	Check(row shared.Row) error
}
