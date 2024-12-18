package impls

type TableStatistics struct {
	RowCount int
}

type ColumnStatistics struct {
	NullCount       int
	DistinctCount   int
	HistogramBounds []any
}
