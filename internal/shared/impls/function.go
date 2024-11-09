package impls

type Function interface {
	Invoke(ctx ExecutionContext, args []any) (any, error)
}
