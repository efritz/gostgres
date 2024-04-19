package nodes

import (
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type Node interface {
	Name() string
	Fields() []shared.Field
	Serialize(w io.Writer, indentationLevel int)
	Optimize()
	AddFilter(filter expressions.Expression)
	AddOrder(order expressions.OrderExpression)
	Filter() expressions.Expression
	Ordering() expressions.OrderExpression

	// TODO: rough implementation
	// https://sourcegraph.com/github.com/postgres/postgres@06286709ee0637ec7376329a5aa026b7682dcfe2/-/blob/src/backend/executor/execAmi.c?L439:59-439:79
	SupportsMarkRestore() bool
	Scanner(ctx scan.ScanContext) (scan.Scanner, error)
}
