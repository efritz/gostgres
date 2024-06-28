package ast

import (
	"github.com/efritz/gostgres/internal/execution/queries"
)

type BuildContext struct {
	Tables TableGetter // TODO - unnecessary if moved to resolver?
}

type Builder interface {
	Build(ctx BuildContext) (queries.Node, error)
}

type ResolverBuilder interface {
	Resolver
	Builder
}
