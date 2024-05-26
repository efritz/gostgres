package protocol

import (
	"github.com/efritz/gostgres/internal/shared/rows"
)

type ResponseWriter interface {
	SendRow(row rows.Row)
	Done()
	Error(err error)
}
