package functions

type Functionspace struct {
	functions map[string]Function
}

func NewDefaultFunctionspace() *Functionspace {
	return &Functionspace{
		functions: DefaultFunctions(),
	}
}

func NewFunctionspace() *Functionspace {
	return &Functionspace{
		functions: map[string]Function{},
	}
}

func (t *Functionspace) GetFunction(name string) (Function, bool) {
	function, ok := t.functions[name]
	return function, ok
}

func (t *Functionspace) SetFunction(name string, function Function) error {
	t.functions[name] = function
	return nil
}
