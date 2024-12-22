package impls

import "github.com/efritz/gostgres/internal/shared/fields"

type TableStatistics struct {
	RowCount         int
	ColumnStatistics []ColumnStatistics
}

type ColumnStatistics struct {
	Field                   fields.Field
	InverseNullProportion   float64
	InverseDistinctFraction float64
	HistogramBounds         []any
}

type IndexStatistics struct {
	RowCount int
}
