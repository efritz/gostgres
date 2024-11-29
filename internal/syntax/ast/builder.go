package ast

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type Resolver interface {
	Resolve(ctx *impls.NodeResolutionContext) error
}

type Builder interface {
	Build() (queries.LogicalNode, error)
}

type BuilderResolver interface {
	Builder
	Resolver
}
