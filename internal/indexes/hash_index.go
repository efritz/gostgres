package indexes

import (
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type hashIndex struct {
	name       string
	tableName  string
	expression expressions.Expression
	entries    map[uint64][]hashItem
}

type hashItem struct {
	tid   int
	value any
}

type hashIndexScanOptions struct {
	expression expressions.Expression
}

var _ Index[hashIndexScanOptions] = &hashIndex{}

func NewHashIndex(name, tableName string, expression expressions.Expression) *hashIndex {
	return &hashIndex{
		name:       name,
		tableName:  tableName,
		expression: expression,
		entries:    map[uint64][]hashItem{},
	}
}

func (i *hashIndex) Unwrap() BaseIndex {
	return i
}

func (i *hashIndex) Filter() expressions.Expression {
	return nil
}

func (i *hashIndex) Description(opts hashIndexScanOptions) string {
	return fmt.Sprintf("hash index scan of %s via %s", i.tableName, i.name)
}

func (i *hashIndex) Condition(opts hashIndexScanOptions) (expr expressions.Expression) {
	if i.expression == nil {
		return nil
	}

	return expressions.NewEquals(i.expression, opts.expression)
}

func (i *hashIndex) Ordering(opts hashIndexScanOptions) expressions.OrderExpression {
	return nil
}

func (i *hashIndex) Insert(row shared.Row) error {
	tid, value, err := i.extractTIDAndValueFromRow(row)
	if err != nil {
		return err
	}

	hash := shared.Hash(value)
	i.entries[hash] = append(i.entries[hash], hashItem{tid, value})
	return nil
}

func (i *hashIndex) Delete(row shared.Row) error {
	tid, value, err := i.extractTIDAndValueFromRow(row)
	if err != nil {
		return err
	}

	hash := shared.Hash(value)
	items := i.entries[hash]

	for j, item := range items {
		if item.tid == tid {
			items[j] = items[len(items)-1]
			i.entries[hash] = items[:len(items)-1]
			break
		}
	}

	return nil
}
