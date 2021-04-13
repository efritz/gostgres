package filters

import "github.com/efritz/gostgres/internal/shared"

type Filter interface {
	Test(row shared.Row) (bool, error)
}
