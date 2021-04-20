package parsing

import "github.com/efritz/gostgres/internal/syntax/tokens"

type Precedence int

const (
	PrecedenceUnknown Precedence = iota
	PrecedenceConditionalOr
	PrecedenceConditionalAnd
	PrecedenceEquality
	PrecedenceComparison
	PrecedenceLike
	PrecedenceBetween
	// PrecedenceIn
	PrecedenceGenericOperator
	PrecedenceIsNotNull
	PrecedenceIsNull
	PrecedenceIs
	PrecedenceAdditive
	PrecedenceMultiplicative
	PrecedenceUnary
	PrecedencePostfix
	PrecedenceAny
)

var precedenceMap = map[tokens.TokenType]Precedence{
	tokens.TokenTypeAnd:                 PrecedenceConditionalAnd,
	tokens.TokenTypeOr:                  PrecedenceConditionalOr,
	tokens.TokenTypeMinus:               PrecedenceAdditive,
	tokens.TokenTypeAsterisk:            PrecedenceMultiplicative,
	tokens.TokenTypeSlash:               PrecedenceMultiplicative,
	tokens.TokenTypePlus:                PrecedenceAdditive,
	tokens.TokenTypeLessThan:            PrecedenceComparison,
	tokens.TokenTypeEquals:              PrecedenceEquality,
	tokens.TokenTypeGreaterThan:         PrecedenceComparison,
	tokens.TokenTypeLessThanOrEqual:     PrecedenceComparison,
	tokens.TokenTypeNotEquals:           PrecedenceEquality,
	tokens.TokenTypeGreaterThanOrEqual:  PrecedenceComparison,
	tokens.TokenTypeIsTrue:              PrecedencePostfix,
	tokens.TokenTypeIsNotTrue:           PrecedencePostfix,
	tokens.TokenTypeIsFalse:             PrecedencePostfix,
	tokens.TokenTypeIsNotFalse:          PrecedencePostfix,
	tokens.TokenTypeIsNull:              PrecedencePostfix,
	tokens.TokenTypeIsNotNull:           PrecedencePostfix,
	tokens.TokenTypeIsUnknown:           PrecedencePostfix,
	tokens.TokenTypeIsNotUnknown:        PrecedencePostfix,
	tokens.TokenTypeConcat:              PrecedenceGenericOperator,
	tokens.TokenTypeIsDistinctFrom:      PrecedenceIs,
	tokens.TokenTypeIsNotDistinctFrom:   PrecedenceIs,
	tokens.TokenTypeLike:                PrecedenceLike,
	tokens.TokenTypeNotLike:             PrecedenceLike,
	tokens.TokenTypeILike:               PrecedenceLike,
	tokens.TokenTypeNotILike:            PrecedenceLike,
	tokens.TokenTypeBetween:             PrecedenceBetween,
	tokens.TokenTypeNotBetween:          PrecedenceBetween,
	tokens.TokenTypeBetweenSymmetric:    PrecedenceBetween,
	tokens.TokenTypeNotBetweenSymmetric: PrecedenceBetween,
}
