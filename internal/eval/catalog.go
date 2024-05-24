package eval

type Catalog[T any] struct {
	entries map[string]T
}

func NewCatalog[T any]() *Catalog[T] {
	return NewCatalogWithEntries[T](nil)
}

func NewCatalogWithEntries[T any](entries map[string]T) *Catalog[T] {
	if entries == nil {
		entries = map[string]T{}
	}

	return &Catalog[T]{
		entries: entries,
	}
}

func (c Catalog[T]) Get(name string) (T, bool) {
	entry, ok := c.entries[name]
	return entry, ok
}

func (c Catalog[T]) Set(name string, entry T) {
	c.entries[name] = entry
}
