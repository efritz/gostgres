package nodes

import (
	"fmt"
	"hash/fnv"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type hashIndex struct {
	name        string
	table       *Table
	expressions []expressions.Expression
	entries     map[int64][]hashItem
}

type hashItem struct {
	tid    int
	values []interface{}
}

type hashIndexScanOptions struct {
	values []interface{}
	// TODO - reconstruct instead of storing explicitly
	expr expressions.Expression
}

func (o hashIndexScanOptions) Condition() expressions.Expression {
	return o.expr
}

var _ Index[hashIndexScanOptions] = &hashIndex{}

func NewHashIndex(name string, table *Table, expressions []expressions.Expression) *hashIndex {
	return &hashIndex{
		name:        name,
		table:       table,
		expressions: expressions,
		entries:     map[int64][]hashItem{},
	}
}

func (i *hashIndex) Name() string {
	return i.name
}

func (i *hashIndex) Filter() expressions.Expression {
	return nil
}

func (i *hashIndex) Ordering() OrderExpression {
	return nil
}

func (i *hashIndex) Insert(row shared.Row) error {
	tid, values, err := i.extractTIDAndValuesFromRow(row)
	if err != nil {
		return err
	}

	hash := i.hash(values)
	i.entries[hash] = append(i.entries[hash], hashItem{tid, values})
	return nil
}

func (i *hashIndex) Delete(row shared.Row) error {
	tid, values, err := i.extractTIDAndValuesFromRow(row)
	if err != nil {
		return err
	}

	hash := i.hash(values)
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

func (i *hashIndex) Scanner(opts hashIndexScanOptions) (tidScanner, error) {
	items := i.entries[i.hash(opts.values)]

	j := 0

	return tidScannerFunc(func() (int, error) {
		if j < len(items) {
			tid := items[j].tid
			j++
			return tid, nil
		}

		return 0, ErrNoRows
	}), nil
}

func (i *hashIndex) extractTIDAndValuesFromRow(row shared.Row) (int, []interface{}, error) {
	tid, ok := extractTID(row)
	if !ok {
		return 0, nil, fmt.Errorf("no tid in row")
	}

	values := []interface{}{}
	for _, expression := range i.expressions {
		value, err := expression.ValueFrom(row)
		if err != nil {
			return 0, nil, err
		}

		values = append(values, value)
	}

	return tid, values, nil
}

func (i *hashIndex) hash(values []interface{}) int64 {
	h := fnv.New64()
	for _, value := range values {
		_, _ = h.Write([]byte(fmt.Sprintf("%v", value)))
	}

	return int64(h.Sum64())
}
