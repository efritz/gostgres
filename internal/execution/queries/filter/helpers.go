package filter

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

func LowerFilter(ctx impls.OptimizationContext, filter impls.Expression, nodes ...queries.Node) {
	for _, expression := range expressions.Conjunctions(filter) {
		for _, node := range nodes {
			lowerFilter(ctx, expression, node)
		}
	}
}

func lowerFilter(ctx impls.OptimizationContext, expression impls.Expression, node queries.Node) {
	expressionFields := expressions.Fields(expression)
	nodeFields := node.Fields()
	outerFields := ctx.OuterFields()
	combinedNodeFields := append(slices.Clone(nodeFields), outerFields...)

	// Ensure that we don't push down filters that references unknown fields
	// Also skip filters that aren't _specific_ to the node and only reference outer fields
	if matchesAllFields(expressionFields, combinedNodeFields) && !matchesAllFields(expressionFields, outerFields) {
		node.AddFilter(ctx, expression)
	}
}

func matchesAllFields(haystack, candidates []fields.Field) bool {
	for _, f := range haystack {
		if _, err := fields.FindMatchingFieldIndex(f, candidates); err != nil {
			return false
		}
	}

	return true
}
