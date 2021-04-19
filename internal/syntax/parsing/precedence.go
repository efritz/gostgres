package parsing

import "github.com/efritz/gostgres/internal/syntax/tokens"

type Precedence int

const (
	PrecedenceUnknown Precedence = iota
	PrecedenceConditionalOr
	PrecedenceConditionalAnd
	PrecedenceEquality
	PrecedenceComparison
	PrecedenceIs
	PrecedenceAdditive
	PrecedenceMultiplicative
	PrecedenceUnary
	PrecedenceAny
)

var precedenceMap = map[tokens.TokenType]Precedence{
	tokens.TokenTypeAnd:                PrecedenceConditionalAnd,
	tokens.TokenTypeIs:                 PrecedenceIs,
	tokens.TokenTypeOr:                 PrecedenceConditionalOr,
	tokens.TokenTypeMinus:              PrecedenceAdditive,
	tokens.TokenTypeAsterisk:           PrecedenceMultiplicative,
	tokens.TokenTypeSlash:              PrecedenceMultiplicative,
	tokens.TokenTypePlus:               PrecedenceAdditive,
	tokens.TokenTypeLessThan:           PrecedenceComparison,
	tokens.TokenTypeEquals:             PrecedenceEquality,
	tokens.TokenTypeGreaterThan:        PrecedenceComparison,
	tokens.TokenTypeLessThanOrEqual:    PrecedenceComparison,
	tokens.TokenTypeNotEquals:          PrecedenceEquality,
	tokens.TokenTypeGreaterThanOrEqual: PrecedenceComparison,
}
