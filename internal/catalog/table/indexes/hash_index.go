package indexes

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/utils"
)

type hashIndex struct {
	name       string
	tableName  string
	expression impls.Expression
	entries    map[uint64][]hashItem
}

type hashItem struct {
	tid   int64
	value any
}

type HashIndexScanOptions struct {
	expression impls.Expression
}

var _ impls.Index[HashIndexScanOptions] = &hashIndex{}

func NewHashIndex(name, tableName string, expression impls.Expression) impls.Index[HashIndexScanOptions] {
	return &hashIndex{
		name:       name,
		tableName:  tableName,
		expression: expression,
		entries:    map[uint64][]hashItem{},
	}
}

func (i *hashIndex) Unwrap() impls.BaseIndex {
	return i
}

func (i *hashIndex) UniqueOn() []fields.Field {
	return nil
}

func (i *hashIndex) Filter() impls.Expression {
	return nil
}

func (i *hashIndex) Name() string {
	return i.name
}

func (i *hashIndex) Description(opts HashIndexScanOptions) string {
	return fmt.Sprintf("hash index scan of %s via %s", i.tableName, i.name)
}

func (i *hashIndex) Condition(opts HashIndexScanOptions) (expr impls.Expression) {
	if i.expression == nil {
		return nil
	}

	return expressions.NewEquals(i.expression, opts.expression)
}

func (i *hashIndex) Ordering(opts HashIndexScanOptions) impls.OrderExpression {
	return nil
}

func (i *hashIndex) Insert(row rows.Row) error {
	tid, value, err := i.extractTIDAndValueFromRow(row)
	if err != nil {
		return err
	}

	hash := utils.Hash(value)
	i.entries[hash] = append(i.entries[hash], hashItem{tid, value})
	return nil
}

func (i *hashIndex) Delete(row rows.Row) error {
	tid, value, err := i.extractTIDAndValueFromRow(row)
	if err != nil {
		return err
	}

	hash := utils.Hash(value)
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

func (i *hashIndex) extractTIDAndValueFromRow(row rows.Row) (int64, any, error) {
	tid, err := row.TID()
	if err != nil {
		return 0, nil, err
	}

	value, err := i.expression.ValueFrom(impls.EmptyContext, row)
	if err != nil {
		return 0, nil, err
	}

	return tid, value, nil
}
