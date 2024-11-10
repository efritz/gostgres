package impls

type Function interface {
	Callable
	Invoke(ctx ExecutionContext, args []any) (any, error)
}
