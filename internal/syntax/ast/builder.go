package ast

import (
	"github.com/efritz/gostgres/internal/execution/queries"
)

type Builder interface {
	Build() (queries.Node, error)
}

type ResolverBuilder interface {
	Resolver
	Builder
}
