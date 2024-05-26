package impls

type Function interface {
	Invoke(ctx Context, args []any) (any, error)
}
