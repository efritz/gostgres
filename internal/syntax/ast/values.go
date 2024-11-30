package ast

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type ValuesBuilder struct {
	Fields      []fields.Field
	Expressions [][]impls.Expression
}

func (b *ValuesBuilder) Resolve(ctx *impls.NodeResolutionContext) error {
	return nil
}

func (b *ValuesBuilder) TableFields() []fields.Field {
	return slices.Clone(b.Fields)
}

func (b *ValuesBuilder) Build() (nodes.LogicalNode, error) {
	return nodes.NewValues(b.Fields, b.Expressions), nil
}
