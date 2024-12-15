package util

import (
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

// TODO - reorganize packages so this isn't necessary
type LogicalNode interface {
	Fields() []fields.Field
	AddFilter(ctx impls.OptimizationContext, filter impls.Expression)
	AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression)
}
