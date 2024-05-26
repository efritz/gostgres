package impls

type Aggregate interface {
	Step(state any, args []any) (any, error)
	Done(state any) (any, error)
}
