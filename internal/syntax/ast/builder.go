package ast

import "github.com/efritz/gostgres/internal/execution/queries"

type Builder interface {
	Resolve(ctx *ResolutionContext) error
	Build(ctx BuildContext) (queries.Node, error)
}
