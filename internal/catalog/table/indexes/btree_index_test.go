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

func TestBTreeIndex(t *testing.T) {
	tid := fields.NewInternalField("authors", rows.TIDName, types.TypeBigInteger)
	id := fields.NewField("authors", "id", types.TypeInteger)
	name := fields.NewField("authors", "name", types.TypeText)
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

	index := NewBTreeIndex("authors_pkey", "authors", false, []impls.ExpressionWithDirection{
		{Expression: expressions.NewNamed(id), Reverse: false},
	})

	for _, row := range allRows {
		require.NoError(t, index.Insert(row))
	}

	ids := func(opts BtreeIndexScanOptions) []int32 {
		scanner, err := index.Scanner(impls.EmptyContext, opts)
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

		return ids
	}

	t.Run("lower-bounded scan", func(t *testing.T) {
		opts := BtreeIndexScanOptions{
			scanDirection: ScanDirectionForward,
			lowerBounds: [][]scanBound{{{
				expression: expressions.NewConstant(int32(750)),
				inclusive:  true,
			}}},
		}

		var expectedIDs []int32
		for i := 750; i <= 1000; i++ {
			expectedIDs = append(expectedIDs, int32(i))
		}
		assert.Equal(t, expectedIDs, ids(opts))
	})

	t.Run("upper-bounded scan", func(t *testing.T) {
		opts := BtreeIndexScanOptions{
			scanDirection: ScanDirectionForward,
			upperBounds: [][]scanBound{{{
				expression: expressions.NewConstant(int32(250)),
				inclusive:  true,
			}}},
		}

		var expectedIDs []int32
		for i := 1; i <= 250; i++ {
			expectedIDs = append(expectedIDs, int32(i))
		}
		assert.Equal(t, expectedIDs, ids(opts))
	})

	t.Run("double-bounded scan", func(t *testing.T) {
		opts := BtreeIndexScanOptions{
			scanDirection: ScanDirectionForward,
			lowerBounds: [][]scanBound{{{
				expression: expressions.NewConstant(int32(250)),
				inclusive:  false,
			}}},
			upperBounds: [][]scanBound{{{
				expression: expressions.NewConstant(int32(750)),
				inclusive:  false,
			}}},
		}

		var expectedIDs []int32
		for i := 251; i < 750; i++ {
			expectedIDs = append(expectedIDs, int32(i))
		}
		assert.Equal(t, expectedIDs, ids(opts))
	})

	t.Run("delete even ids", func(t *testing.T) {
		for k, row := range allRows {
			if k%2 == 0 {
				require.NoError(t, index.Delete(row))
			}
		}
	})

	t.Run("unbounded scan", func(t *testing.T) {
		opts := BtreeIndexScanOptions{
			scanDirection: ScanDirectionForward,
		}

		var expectedIDs []int32
		for i := 1; i <= 1000; i++ {
			if i%2 != 0 {
				expectedIDs = append(expectedIDs, int32(i))
			}
		}
		assert.Equal(t, expectedIDs, ids(opts))
	})
}
