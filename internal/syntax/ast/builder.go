package ast

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type Resolver interface {
	Resolve(ctx *impls.NodeResolutionContext) error
}

type Builder interface {
	Build() (nodes.LogicalNode, error)
}

type BuilderResolver interface {
	Builder
	Resolver
}
