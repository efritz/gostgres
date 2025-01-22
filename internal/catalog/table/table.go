package table

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/rows"
	"golang.org/x/exp/slices"
)

type table struct {
	name            string
	fields          []impls.TableField
	rows            map[int64]rows.Row
	primaryKey      impls.BaseIndex
	indexes         []impls.BaseIndex
	constraints     []impls.Constraint
	tableStatistics impls.TableStatistics
}

var _ impls.Table = &table{}

func NewTable(name string, nonInternalFields []impls.TableField) impls.Table {
	tableFields := []impls.TableField{
		impls.NewTableFieldFromField(fields.TIDField.WithRelationName(name)),
	}
	for _, field := range nonInternalFields {
		tableFields = append(tableFields, field.WithRelationName(name))
	}

	return &table{
		name:            name,
		fields:          tableFields,
		rows:            map[int64]rows.Row{},
		tableStatistics: impls.TableStatistics{},
	}
}

func (t *table) Name() string {
	return t.name
}

func (t *table) Indexes() []impls.BaseIndex {
	if t.primaryKey != nil {
		return append([]impls.BaseIndex{t.primaryKey}, t.indexes...)
	}

	return t.indexes
}

func (t *table) Fields() []impls.TableField {
	return slices.Clone(t.fields)
}

func (t *table) TIDs() []int64 {
	tids := make([]int64, 0, len(t.rows))
	for tid := range t.rows {
		tids = append(tids, tid)
	}
	slices.Sort(tids)

	return tids
}

func (t *table) Row(tid int64) (rows.Row, bool) {
	row, ok := t.rows[tid]
	return row, ok
}

func (t *table) SetPrimaryKey(index impls.BaseIndex) error {
	if t.primaryKey != nil {
		return fmt.Errorf("primary key already set")
	}

	for _, row := range t.rows {
		if err := index.Insert(row); err != nil {
			return err
		}
	}

	t.primaryKey = index
	return nil
}

func (t *table) AddIndex(index impls.BaseIndex) error {
	for _, row := range t.rows {
		if err := index.Insert(row); err != nil {
			return err
		}
	}

	t.indexes = append(t.indexes, index)
	return nil
}

func (t *table) AddConstraint(ctx impls.ExecutionContext, constraint impls.Constraint) error {
	for _, row := range t.rows {
		if err := constraint.Check(ctx, row); err != nil {
			return err
		}
	}

	t.constraints = append(t.constraints, constraint)
	return nil
}

var tid = int64(0)

func (t *table) Insert(ctx impls.ExecutionContext, row rows.Row) (_ rows.Row, err error) {
	tid++
	id := tid

	var fields []fields.Field
	for _, field := range t.fields {
		fields = append(fields, field.Field)
	}

	newRow, err := rows.NewRow(fields, append([]any{id}, row.Values...))
	if err != nil {
		return rows.Row{}, err
	}

	for _, constraint := range t.constraints {
		if err := constraint.Check(ctx, newRow); err != nil {
			return rows.Row{}, err
		}
	}

	for _, index := range t.Indexes() {
		if err := index.Insert(newRow); err != nil {
			return rows.Row{}, err
		}
	}

	t.rows[id] = newRow
	return newRow, nil
}

func (t *table) Delete(row rows.Row) (rows.Row, bool, error) {
	tid, err := row.TID()
	if err != nil {
		return rows.Row{}, false, err
	}

	fullRow, ok := t.rows[tid]
	if !ok {
		return rows.Row{}, false, nil
	}

	for _, index := range t.Indexes() {
		if err := index.Delete(fullRow); err != nil {
			return rows.Row{}, false, err
		}
	}

	delete(t.rows, tid)
	return fullRow, true, nil
}

func (t *table) Statistics() impls.TableStatistics {
	return t.tableStatistics
}

func (t *table) Analyze() error {
	tableStats := impls.TableStatistics{
		RowCount:         len(t.rows),
		ColumnStatistics: make([]impls.ColumnStatistics, len(t.fields)),
	}

	for columnIndex := range t.fields {
		tableStats.ColumnStatistics[columnIndex] = t.analyzeColumn(columnIndex)
	}

	for _, index := range t.Indexes() {
		if err := index.Analyze(); err != nil {
			return err
		}
	}

	t.tableStatistics = tableStats
	return nil
}

func (t *table) analyzeColumn(columnIndex int) impls.ColumnStatistics {
	nullCount := 0
	countsByValue := map[any]int{}

	for _, row := range t.rows {
		value := row.Values[columnIndex]

		if value == nil {
			nullCount++
		} else {
			countsByValue[value] = countsByValue[value] + 1
		}
	}

	mostCommonValues := t.calculateMostCommonValues(countsByValue)

	commonValueSet := map[any]struct{}{}
	for _, mcv := range mostCommonValues {
		commonValueSet[mcv.Value] = struct{}{}
	}
	var nonNilNonCommonValues []any
	for value, count := range countsByValue {
		if _, ok := commonValueSet[value]; !ok {
			for i := 0; i < count; i++ {
				nonNilNonCommonValues = append(nonNilNonCommonValues, value)
			}
		}
	}

	nullFraction := 0.0
	distinctFraction := 0.0
	if total := len(t.rows); total > 0 {
		nullFraction = float64(nullCount) / float64(total)

		if nonNullCount := total - nullCount; nonNullCount > 0 {
			distinctFraction = float64(len(countsByValue)) / float64(nonNullCount)
		}
	}

	var minValue, maxValue any
	for value := range countsByValue {
		if minValue == nil || ordering.CompareValues(value, minValue) == ordering.OrderTypeBefore {
			minValue = value
		}

		if maxValue == nil || ordering.CompareValues(value, maxValue) == ordering.OrderTypeAfter {
			maxValue = value
		}
	}

	return impls.ColumnStatistics{
		Field:            t.fields[columnIndex].Field,
		NullFraction:     nullFraction,
		DistinctFraction: distinctFraction,
		MinValue:         minValue,
		MaxValue:         maxValue,
		MostCommonValues: mostCommonValues,
		HistogramBounds:  t.calculateHistogramBounds(nonNilNonCommonValues),
	}
}

const maxMostCommonValues = 20
const minimumCommonFrequency = 0.001

func (t *table) calculateMostCommonValues(countsByValue map[any]int) []impls.MostCommonValue {
	var mostCommonValues []impls.MostCommonValue
	for value, count := range countsByValue {
		if float64(count)/float64(len(t.rows)) >= minimumCommonFrequency {
			mostCommonValues = append(mostCommonValues, impls.MostCommonValue{Value: value, Frequency: float64(count) / float64(len(t.rows))})
		}
	}

	slices.SortFunc(mostCommonValues, func(a, b impls.MostCommonValue) int {
		if a.Frequency < b.Frequency {
			return 1
		} else if a.Frequency > b.Frequency {
			return -1
		}

		return 0
	})

	if len(mostCommonValues) > maxMostCommonValues {
		mostCommonValues = mostCommonValues[:maxMostCommonValues]
	}

	return mostCommonValues
}

const numHistogramBuckets = 100

func (t *table) calculateHistogramBounds(nonNilNonCommonValues []any) []any {
	if len(nonNilNonCommonValues) < 2 || ordering.CompareValues(nonNilNonCommonValues[0], nonNilNonCommonValues[1]) == ordering.OrderTypeIncomparable {
		return nil
	}

	slices.SortFunc(nonNilNonCommonValues, func(a, b any) int {
		switch ordering.CompareValues(a, b) {
		case ordering.OrderTypeBefore:
			return -1
		case ordering.OrderTypeAfter:
			return 1
		default:
			return 0
		}
	})

	if len(nonNilNonCommonValues) <= numHistogramBuckets {
		return nonNilNonCommonValues
	}

	bounds := make([]any, 0, numHistogramBuckets-1)
	for i := 1; i < numHistogramBuckets; i++ {
		bounds = append(bounds, nonNilNonCommonValues[(i*len(nonNilNonCommonValues))/numHistogramBuckets])
	}

	return bounds
}
