package expressions

import "github.com/efritz/gostgres/internal/shared"

type Expression interface {
	ValueFrom(row shared.Row) (interface{}, error)
}
