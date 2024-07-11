package ast

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/syntax/ast/context"
)

type Resolver interface {
	Resolve(ctx *context.ResolveContext) error
}

type Builder interface {
	Build() (queries.Node, error)
}

type BuilderResolver interface {
	Builder
	Resolver
}
