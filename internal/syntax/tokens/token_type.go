package tokens

type TokenType int

const (
	TokenTypeInvalid TokenType = iota
	TokenTypeEOF
	TokenTypeKeyword
	TokenTypeWhitespace
	TokenTypeIdent
	TokenTypeNumber
	TokenTypeString

	//
	// Keywords

	TokenTypeAll
	TokenTypeAnd
	TokenTypeAs
	TokenTypeAscending
	TokenTypeBetween
	TokenTypeBy
	TokenTypeDelete
	TokenTypeDescending
	TokenTypeDistinct
	TokenTypeExcept
	TokenTypeFalse
	TokenTypeFrom
	TokenTypeILike
	TokenTypeInsert
	TokenTypeIntersect
	TokenTypeInto
	TokenTypeIs
	TokenTypeJoin
	TokenTypeLike
	TokenTypeLimit
	TokenTypeNot
	TokenTypeNull
	TokenTypeOffset
	TokenTypeOn
	TokenTypeOr
	TokenTypeOrder
	TokenTypeReturning
	TokenTypeSelect
	TokenTypeSet
	TokenTypeSymmetric
	TokenTypeTrue
	TokenTypeUnion
	TokenTypeKwUnknown
	TokenTypeUpdate
	TokenTypeUsing
	TokenTypeValues
	TokenTypeWhere

	//
	// Single-character operators

	TokenTypeMinus
	TokenTypeComma
	TokenTypeSemicolon
	TokenTypeDot
	TokenTypeLeftParen
	TokenTypeRightParen
	TokenTypeAsterisk
	TokenTypeSlash
	TokenTypePlus
	TokenTypeLessThan
	TokenTypeEquals
	TokenTypeGreaterThan

	//
	// Multiple-character operators

	TokenTypeLessThanOrEqual
	TokenTypeNotEquals
	TokenTypeGreaterThanOrEqual
	TokenTypeConcat

	//
	// Multiple-keyword operators

	TokenTypeBetweenSymmetric
	TokenTypeIsDistinctFrom
	TokenTypeIsFalse
	TokenTypeIsNotDistinctFrom
	TokenTypeIsNotFalse
	TokenTypeIsNotNull
	TokenTypeIsNotTrue
	TokenTypeIsNotUnknown
	TokenTypeIsNull
	TokenTypeIsTrue
	TokenTypeIsUnknown
	TokenTypeNotBetween
	TokenTypeNotBetweenSymmetric
	TokenTypeNotLike
	TokenTypeNotILike

	TokenTypeUnknown
)
