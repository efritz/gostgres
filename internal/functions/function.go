package functions

type Function func(ctx FunctionContext, args []any) (any, error)
