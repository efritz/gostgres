package impls

type Aggregate interface {
	Callable
	Step(ctx ExecutionContext, state any, args []any) (any, error)
	Done(ctx ExecutionContext, state any) (any, error)
}
