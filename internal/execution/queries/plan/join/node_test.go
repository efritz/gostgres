package join

import (
	"testing"

	"github.com/efritz/gostgres/internal/execution/queries/nodes/join"
	"github.com/stretchr/testify/assert"
)

func TestConvertRightJoinsToLeftJoins(t *testing.T) {
	a := NewJoinLeafNode(newMockLogicalNode("a"))
	b := NewJoinLeafNode(newMockLogicalNode("b"))
	c := NewJoinLeafNode(newMockLogicalNode("c"))
	d := NewJoinLeafNode(newMockLogicalNode("d"))

	makeJoin := func(left, right JoinNode, joinType join.JoinType) JoinNode {
		return NewJoinInternalNode(left, right, JoinOperator{JoinType: joinType})
	}

	testCases := []struct {
		name     string
		input    JoinNode
		expected JoinNode
	}{
		{
			name: "no right joins",
			// FROM a JOIN b
			// FROM a JOIN b
			input:    makeJoin(a, b, join.JoinTypeInner),
			expected: makeJoin(a, b, join.JoinTypeInner),
		},
		{
			name: "right join to single relation",
			// FROM ((a JOIN b) RIGHT JOIN (c)) JOIN d
			// FROM ((c)  LEFT JOIN (a JOIN b)) JOIN d
			input:    makeJoin(makeJoin(makeJoin(a, b, join.JoinTypeInner), c, join.JoinTypeRightOuter), d, join.JoinTypeInner),
			expected: makeJoin(makeJoin(c, makeJoin(a, b, join.JoinTypeInner), join.JoinTypeLeftOuter), d, join.JoinTypeInner),
		},
		{
			name: "right join to multiple relations",
			// FROM (a JOIN b) RIGHT JOIN (c JOIN d)
			// FROM (c JOIN d)  LEFT JOIN (a JOIN b)
			input:    makeJoin(makeJoin(a, b, join.JoinTypeInner), makeJoin(c, d, join.JoinTypeInner), join.JoinTypeRightOuter),
			expected: makeJoin(makeJoin(c, d, join.JoinTypeInner), makeJoin(a, b, join.JoinTypeInner), join.JoinTypeLeftOuter),
		},
		{
			name: "multiple right joins",
			// FROM ((a RIGHT JOIN b) RIGHT JOIN c) JOIN d
			// FROM (c  LEFT JOIN (b  LEFT JOIN a)) JOIN d
			input:    makeJoin(makeJoin(makeJoin(a, b, join.JoinTypeRightOuter), c, join.JoinTypeRightOuter), d, join.JoinTypeInner),
			expected: makeJoin(makeJoin(c, makeJoin(b, a, join.JoinTypeLeftOuter), join.JoinTypeLeftOuter), d, join.JoinTypeInner),
		},
		{
			name: "nested right joins",
			// FROM (a RIGHT JOIN (b JOIN (c RIGHT JOIN d)))
			// FROM ((b JOIN (d  LEFT JOIN c))  LEFT JOIN a)
			input:    makeJoin(a, makeJoin(b, makeJoin(c, d, join.JoinTypeRightOuter), join.JoinTypeInner), join.JoinTypeRightOuter),
			expected: makeJoin(makeJoin(b, makeJoin(d, c, join.JoinTypeLeftOuter), join.JoinTypeInner), a, join.JoinTypeLeftOuter),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, testCase.input.ConvertRightJoinsToLeftJoins())
		})
	}
}
