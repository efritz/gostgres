package fields

import (
	"testing"

	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindMatchingFieldIndex(t *testing.T) {
	t.Run("matching qualified field", func(t *testing.T) {
		index, err := FindMatchingFieldIndex(
			NewField("t", "b", types.TypeText),
			[]Field{
				NewField("t", "a", types.TypeText),
				NewField("t", "b", types.TypeText),
				NewField("t", "c", types.TypeText),
			},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, index)
	})

	t.Run("matching unqualified field", func(t *testing.T) {
		index, err := FindMatchingFieldIndex(
			NewField("", "b", types.TypeText),
			[]Field{
				NewField("t", "a", types.TypeText),
				NewField("t", "b", types.TypeText),
				NewField("t", "c", types.TypeText),
			},
		)
		require.NoError(t, err)
		assert.Equal(t, 1, index)
	})

	t.Run("ambiguous field", func(t *testing.T) {
		_, err := FindMatchingFieldIndex(
			NewField("", "b", types.TypeText),
			[]Field{
				NewField("t1", "a", types.TypeText),
				NewField("t1", "b", types.TypeText),
				NewField("t2", "b", types.TypeText),
				NewField("t2", "c", types.TypeText),
			},
		)
		require.ErrorContains(t, err, `ambiguous field "b"`)
	})

	t.Run("unknown field", func(t *testing.T) {
		_, err := FindMatchingFieldIndex(
			NewField("t", "d", types.TypeText),
			[]Field{
				NewField("t", "a", types.TypeText),
				NewField("t", "b", types.TypeText),
				NewField("t", "c", types.TypeText),
			},
		)
		require.ErrorContains(t, err, `unknown field "t"."d"`)
	})
}
