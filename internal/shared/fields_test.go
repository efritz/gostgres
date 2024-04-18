package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindMatchingFieldIndex(t *testing.T) {
	t.Run("matching qualified field", func(t *testing.T) {
		index, err := FindMatchingFieldIndex(
			NewField("t", "b", TypeKindText, false),
			[]Field{
				NewField("t", "a", TypeKindText, false),
				NewField("t", "b", TypeKindText, false),
				NewField("t", "c", TypeKindText, false),
			},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, index)
	})

	t.Run("matching unqualified field", func(t *testing.T) {
		index, err := FindMatchingFieldIndex(
			NewField("", "b", TypeKindText, false),
			[]Field{
				NewField("t", "a", TypeKindText, false),
				NewField("t", "b", TypeKindText, false),
				NewField("t", "c", TypeKindText, false),
			},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, index)
	})

	t.Run("ambiguous field", func(t *testing.T) {
		_, err := FindMatchingFieldIndex(
			NewField("", "b", TypeKindText, false),
			[]Field{
				NewField("t1", "a", TypeKindText, false),
				NewField("t1", "b", TypeKindText, false),
				NewField("t2", "b", TypeKindText, false),
				NewField("t2", "c", TypeKindText, false),
			},
		)
		require.ErrorContains(t, err, "ambiguous field b")
	})

	t.Run("unknown field", func(t *testing.T) {
		_, err := FindMatchingFieldIndex(
			NewField("t", "d", TypeKindText, false),
			[]Field{
				NewField("t", "a", TypeKindText, false),
				NewField("t", "b", TypeKindText, false),
				NewField("t", "c", TypeKindText, false),
			},
		)
		require.ErrorContains(t, err, "unknown field d")
	})
}
