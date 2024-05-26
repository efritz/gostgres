package queries

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type Node interface {
	Name() string
	Fields() []shared.Field
	Serialize(w serialization.IndentWriter)
	AddFilter(filter types.Expression)
	AddOrder(order expressions.OrderExpression)
	Optimize()
	Filter() types.Expression
	Ordering() expressions.OrderExpression

	// TODO: rough implementation
	// https://sourcegraph.com/github.com/postgres/postgres@06286709ee0637ec7376329a5aa026b7682dcfe2/-/blob/src/backend/executor/execAmi.c?L439:59-439:79
	SupportsMarkRestore() bool
	Scanner(ctx types.Context) (scan.Scanner, error)
}
