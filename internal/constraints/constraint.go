package constraints

import (
	"github.com/efritz/gostgres/internal/table"
)

type Constraint interface {
	table.Constraint
}
