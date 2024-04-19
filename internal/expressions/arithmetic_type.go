package expressions

type ArithmeticType int

const (
	ArithmeticTypeUnknown ArithmeticType = iota
	ArithmeticTypeAddition
	ArithmeticTypeSubtraction
	ArithmeticTypeMultiplication
	ArithmeticTypeDivision
)

var arithmeticTypeOperators = map[ArithmeticType]string{
	ArithmeticTypeAddition:       "+",
	ArithmeticTypeSubtraction:    "-",
	ArithmeticTypeMultiplication: "*",
	ArithmeticTypeDivision:       "/",
}

func ArithmeticTypeFromString(operator string) ArithmeticType {
	for ct, op := range arithmeticTypeOperators {
		if op == operator {
			return ct
		}
	}

	return ArithmeticTypeUnknown
}

func (at ArithmeticType) String() string {
	if operator, ok := arithmeticTypeOperators[at]; ok {
		return operator
	}

	return "unknown"
}

func (at ArithmeticType) IsSymmetric() bool {
	switch at {
	case ArithmeticTypeAddition, ArithmeticTypeMultiplication:
		return true

	default:
		return false
	}
}
