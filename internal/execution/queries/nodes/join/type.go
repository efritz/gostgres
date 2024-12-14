package join

type JoinType int

const (
	JoinTypeUnknown JoinType = iota
	JoinTypeInner
	JoinTypeLeftOuter
	JoinTypeRightOuter
	JoinTypeFullOuter
	JoinTypeCross
)

func (t JoinType) String() string {
	switch t {
	case JoinTypeInner:
		return "INNER JOIN"
	case JoinTypeLeftOuter:
		return "LEFT OUTER JOIN"
	case JoinTypeRightOuter:
		return "RIGHT OUTER JOIN"
	case JoinTypeFullOuter:
		return "FULL OUTER JOIN"
	case JoinTypeCross:
		return "CROSS JOIN"
	}

	return "UNKNOWN JOIN"
}
