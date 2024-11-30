package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type Node interface {
	serialization.Serializable

	Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error)
}
