package impls

import "github.com/efritz/gostgres/internal/shared/fields"

type RelationStatistics struct {
	RowCount         int
	ColumnStatistics []ColumnStatistics
}

type TableStatistics = RelationStatistics

type ColumnStatistics struct {
	Field            fields.Field
	NullFraction     float64
	DistinctFraction float64
	MinValue         any
	MaxValue         any
	MostCommonValues []MostCommonValue
	HistogramBounds  []any
}

type MostCommonValue struct {
	Value     any
	Frequency float64
}

type IndexStatistics struct {
	RowCount int
	// TODO - separate column stats here?
}
