package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefineStringType(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		typ, value, ok := TypeText.Refine("foobar")
		require.True(t, ok)
		assert.Equal(t, TypeText, typ)
		assert.Equal(t, "foobar", value)
	})

	t.Run("string-formatted date", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)

		typ, value, ok := TypeTimestampTz.Refine(now.Format(dateFormat))
		require.True(t, ok)
		assert.Equal(t, TypeTimestampTz, typ)

		timeValue, ok := value.(time.Time)
		require.True(t, ok)
		assert.True(t, now.Truncate(time.Hour*24).Equal(timeValue))
	})

	t.Run("string-formatted timestamp", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)

		typ, value, ok := TypeTimestampTz.Refine(now.Format(timestampFormat))
		require.True(t, ok)
		assert.Equal(t, TypeTimestampTz, typ)

		timeValue, ok := value.(time.Time)
		require.True(t, ok)
		assert.True(t, now.Equal(timeValue))
	})

	t.Run("non-string", func(t *testing.T) {
		_, _, ok := TypeText.Refine(123)
		require.False(t, ok)
	})
}
