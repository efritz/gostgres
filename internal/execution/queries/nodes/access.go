package nodes

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type AccessStrategy interface {
	Serialize(w serialization.IndentWriter)
	Filter() impls.Expression
	Ordering() impls.OrderExpression
	Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error)
}

type accessNode struct {
	table    impls.Table
	strategy AccessStrategy
}

func NewAccess(table impls.Table, strategy AccessStrategy) Node {
	return &accessNode{table, strategy}
}

func (n *accessNode) Serialize(w serialization.IndentWriter) {
	n.strategy.Serialize(w)
}

func (n *accessNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	return n.strategy.Scanner(ctx)
}
