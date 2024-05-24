package filter

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared"
)

func LowerFilter(filter expressions.Expression, nodes ...queries.Node) {
	for _, expression := range expressions.Conjunctions(filter) {
		missing := make([]bool, len(nodes))
		for _, field := range expressions.Fields(expression) {
			for i, node := range nodes {
				if _, err := shared.FindMatchingFieldIndex(field, node.Fields()); err != nil {
					missing[i] = true
				}
			}
		}

		for i, missing := range missing {
			if !missing {
				nodes[i].AddFilter(expression)
			}
		}
	}
}
