package queries

import (
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type Node interface {
	serialization.Serializable

	Name() string
	Fields() []fields.Field
	AddFilter(filter impls.Expression)
	AddOrder(order impls.OrderExpression)
	Optimize()
	Filter() impls.Expression
	Ordering() impls.OrderExpression

	// TODO: rough implementation
	// https://sourcegraph.com/github.com/postgres/postgres@06286709ee0637ec7376329a5aa026b7682dcfe2/-/blob/src/backend/executor/execAmi.c?L439:59-439:79
	SupportsMarkRestore() bool
	Scanner(ctx impls.Context) (scan.RowScanner, error)
}
