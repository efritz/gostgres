package indexes

import (
	"fmt"
	"testing"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashIndex(t *testing.T) {
	tid := fields.NewField("authors", "tid", types.TypeBigInteger, fields.InternalFieldTid)
	id := fields.NewField("authors", "id", types.TypeInteger, fields.NonInternalField)
	name := fields.NewField("authors", "name", types.TypeText, fields.NonInternalField)
	fields := []fields.Field{tid, id, name}

	allRows := map[int64]rows.Row{}
	for i := 1; i <= 1000; i++ {
		tid := int64(i)

		values := []any{
			tid,                       // tid
			int32(i),                  // id
			fmt.Sprintf("name-%d", i), // name
		}

		row, err := rows.NewRow(fields, values)
		require.NoError(t, err)
		allRows[tid] = row
	}

	index := NewHashIndex("authors_name", "authors", expressions.NewNamed(name))

	for _, row := range allRows {
		require.NoError(t, index.Insert(row))
	}

	t.Run("find particular value", func(t *testing.T) {
		opts := HashIndexScanOptions{
			expression: expressions.NewConstant("name-42"),
		}

		scanner, err := index.Scanner(impls.EmptyExecutionContext, opts)
		require.NoError(t, err)

		var ids []int32
		for {
			tid, err := scanner.Scan()
			if err != nil {
				if err == scan.ErrNoRows {
					break
				}

				require.NoError(t, err)
			}

			ids = append(ids, allRows[tid].Values[1].(int32))
		}

		assert.Equal(t, []int32{42}, ids)
	})
}
