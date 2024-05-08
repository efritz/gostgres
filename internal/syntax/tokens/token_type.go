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

	TokenTypeAdd
	TokenTypeAll
	TokenTypeAlter
	TokenTypeAnd
	TokenTypeAs
	TokenTypeAscending
	TokenTypeBetween
	TokenTypeBy
	TokenTypeCheck
	TokenTypeConstraint
	TokenTypeCreate
	TokenTypeDefault
	TokenTypeDelete
	TokenTypeDescending
	TokenTypeDistinct
	TokenTypeExcept
	TokenTypeExplain
	TokenTypeFalse
	TokenTypeFrom
	TokenTypeILike
	TokenTypeIndex
	TokenTypeInsert
	TokenTypeIntersect
	TokenTypeInto
	TokenTypeIs
	TokenTypeJoin
	TokenTypeKey
	TokenTypeKwUnknown
	TokenTypeLike
	TokenTypeLimit
	TokenTypeNot
	TokenTypeNull
	TokenTypeOffset
	TokenTypeOn
	TokenTypeOr
	TokenTypeOrder
	TokenTypePrimary
	TokenTypeReturning
	TokenTypeSelect
	TokenTypeSet
	TokenTypeSymmetric
	TokenTypeTable
	TokenTypeTrue
	TokenTypeUnion
	TokenTypeUnique
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
	TokenTypeNotILike
	TokenTypeNotLike
	TokenTypeNotNull
	TokenTypePrimaryKey

	TokenTypeUnknown
)
