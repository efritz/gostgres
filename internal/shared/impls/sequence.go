package impls

type Sequence interface {
	Name() string
	Next() (int64, error)
	Set(value int64) error
	Value() int64
}
