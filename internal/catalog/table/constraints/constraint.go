package constraints

import "github.com/efritz/gostgres/internal/catalog/table"

type Constraint interface {
	table.Constraint
}
