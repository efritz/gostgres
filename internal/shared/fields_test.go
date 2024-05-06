package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindMatchingFieldIndex(t *testing.T) {
	t.Run("matching qualified field", func(t *testing.T) {
		index, err := FindMatchingFieldIndex(
			NewField("t", "b", TypeText),
			[]Field{
				NewField("t", "a", TypeText),
				NewField("t", "b", TypeText),
				NewField("t", "c", TypeText),
			},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, index)
	})

	t.Run("matching unqualified field", func(t *testing.T) {
		index, err := FindMatchingFieldIndex(
			NewField("", "b", TypeText),
			[]Field{
				NewField("t", "a", TypeText),
				NewField("t", "b", TypeText),
				NewField("t", "c", TypeText),
			},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, index)
	})

	t.Run("ambiguous field", func(t *testing.T) {
		_, err := FindMatchingFieldIndex(
			NewField("", "b", TypeText),
			[]Field{
				NewField("t1", "a", TypeText),
				NewField("t1", "b", TypeText),
				NewField("t2", "b", TypeText),
				NewField("t2", "c", TypeText),
			},
		)
		require.ErrorContains(t, err, `ambiguous field "b"`)
	})

	t.Run("unknown field", func(t *testing.T) {
		_, err := FindMatchingFieldIndex(
			NewField("t", "d", TypeText),
			[]Field{
				NewField("t", "a", TypeText),
				NewField("t", "b", TypeText),
				NewField("t", "c", TypeText),
			},
		)
		require.ErrorContains(t, err, `unknown field "t"."d"`)
	})
}
