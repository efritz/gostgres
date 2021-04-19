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

	TokenTypeAnd
	TokenTypeAs
	TokenTypeAscending
	TokenTypeBy
	TokenTypeDelete
	TokenTypeDescending
	TokenTypeFalse
	TokenTypeFrom
	TokenTypeInsert
	TokenTypeInto
	TokenTypeIs
	TokenTypeIsNull
	TokenTypeJoin
	TokenTypeLimit
	TokenTypeNot
	TokenTypeNotNull
	TokenTypeNull
	TokenTypeOffset
	TokenTypeOn
	TokenTypeOr
	TokenTypeOrder
	TokenTypeReturning
	TokenTypeSelect
	TokenTypeSet
	TokenTypeTrue
	TokenTypeUnion
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

	TokenTypeUnknown
)
