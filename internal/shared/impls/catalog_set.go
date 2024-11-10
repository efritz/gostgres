package impls

import "github.com/efritz/gostgres/internal/catalog"

type CatalogSet struct {
	Tables     *catalog.Catalog[Table]
	Sequences  *catalog.Catalog[Sequence]
	Functions  *catalog.Catalog[Function]
	Aggregates *catalog.Catalog[Aggregate]
}

func NewCatalogEmptySet() CatalogSet {
	return NewCatalogSet(
		catalog.NewCatalog[Table](),
		catalog.NewCatalog[Sequence](),
		catalog.NewCatalog[Function](),
		catalog.NewCatalog[Aggregate](),
	)
}

func NewCatalogSet(
	tables *catalog.Catalog[Table],
	sequences *catalog.Catalog[Sequence],
	functions *catalog.Catalog[Function],
	aggregates *catalog.Catalog[Aggregate],
) CatalogSet {
	return CatalogSet{
		Tables:     tables,
		Sequences:  sequences,
		Functions:  functions,
		Aggregates: aggregates,
	}
}
