package join

import (
	"fmt"
	"testing"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/join"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/types"
)

func TestEnumerateJoinTrees(t *testing.T) {
	forest := EnumerateValidJoinTrees(
		NewJoinInternalNode(
			NewJoinInternalNode(
				NewJoinLeafNode(newMockLogicalNode("a")),
				NewJoinLeafNode(newMockLogicalNode("b")),
				JoinOperator{
					JoinType: join.JoinTypeInner,
					Condition: expressions.NewEquals(
						expressions.NewNamed(fields.NewField("b", "bar", types.TypeUnknown, fields.NonInternalField)),
						expressions.NewNamed(fields.NewField("a", "foo", types.TypeUnknown, fields.NonInternalField)),
					),
				},
			),
			NewJoinLeafNode(newMockLogicalNode("c")),
			JoinOperator{
				JoinType: join.JoinTypeFullOuter,
				Condition: expressions.NewEquals(
					expressions.NewNamed(fields.NewField("c", "foo", types.TypeUnknown, fields.NonInternalField)),
					expressions.NewNamed(fields.NewField("b", "bar", types.TypeUnknown, fields.NonInternalField)),
				),
			},
		),
	)

	unique := map[string]any{}
	for _, tree := range forest {
		fmt.Printf("%s\n", tree)
		unique[tree.String()] = struct{}{}
	}

	fmt.Printf("> %d (%d unique)\n", len(forest), len(unique))
}
