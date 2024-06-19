package ast

import "github.com/efritz/gostgres/internal/execution/queries"

type Builder interface {
	Build(ctx BuildContext) (queries.Node, error)
}
