package nodes

import (
	"fmt"
	"hash/fnv"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type hashIndex struct {
	name       string
	table      *Table
	expression expressions.Expression
	entries    map[uint64][]hashItem
}

type hashItem struct {
	tid   int
	value interface{}
}

type hashIndexScanOptions struct {
	expression expressions.Expression
}

var _ Index[hashIndexScanOptions] = &hashIndex{}

func NewHashIndex(name string, table *Table, expression expressions.Expression) *hashIndex {
	return &hashIndex{
		name:       name,
		table:      table,
		expression: expression,
		entries:    map[uint64][]hashItem{},
	}
}

func (i *hashIndex) Name() string {
	return i.name
}

func (i *hashIndex) Filter() expressions.Expression {
	return nil
}

func (i *hashIndex) Condition(opts hashIndexScanOptions) (expr expressions.Expression) {
	if i.expression == nil {
		return nil
	}

	return expressions.NewEquals(i.expression, opts.expression)
}

func (i *hashIndex) Ordering(opts hashIndexScanOptions) OrderExpression {
	return nil
}

func (i *hashIndex) Insert(row shared.Row) error {
	tid, value, err := i.extractTIDAndValueFromRow(row)
	if err != nil {
		return err
	}

	hash := hash(value)
	i.entries[hash] = append(i.entries[hash], hashItem{tid, value})
	return nil
}

func (i *hashIndex) Delete(row shared.Row) error {
	tid, value, err := i.extractTIDAndValueFromRow(row)
	if err != nil {
		return err
	}

	hash := hash(value)
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

func (i *hashIndex) Scanner(ctx ScanContext, opts hashIndexScanOptions) (tidScanner, error) {
	value, err := ctx.Evaluate(opts.expression, shared.Row{})
	if err != nil {
		return nil, err
	}

	items := i.entries[hash(value)]

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

func (i *hashIndex) extractTIDAndValueFromRow(row shared.Row) (int, interface{}, error) {
	tid, ok := extractTID(row)
	if !ok {
		return 0, nil, fmt.Errorf("no tid in row")
	}

	value, err := i.expression.ValueFrom(row)
	if err != nil {
		return 0, nil, err
	}

	return tid, value, nil
}

func hash(value interface{}) uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(fmt.Sprintf("%v", value)))
	return h.Sum64()
}
