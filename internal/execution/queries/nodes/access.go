package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type AccessStrategy interface {
	Serialize(w serialization.IndentWriter)
	Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error)
}

type accessNode struct {
	AccessStrategy
}

func NewAccess(strategy AccessStrategy) Node {
	return &accessNode{
		AccessStrategy: strategy,
	}
}
