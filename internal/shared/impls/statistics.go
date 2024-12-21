package impls

type TableStatistics struct {
	RowCount int
}

type ColumnStatistics struct {
	InverseNullProportion   float64
	InverseDistinctFraction float64
	HistogramBounds         []any
}

type IndexStatistics struct {
	RowCount int
}
