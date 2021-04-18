package syntax

type Token struct {
	Type   TokenType
	Offset int
	Text   string
}

var InvalidToken = Token{
	Type: TokenTypeInvalid,
}

type TokenType int

const (
	TokenTypeInvalid TokenType = iota
	TokenTypeEOF
	TokenTypeKeyword
	TokenTypeWhitespace
	TokenTypeIdent
	TokenTypeNumber
	TokenTypeString

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
	TokenTypeUpdate
	TokenTypeValues
	TokenTypeWhere

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

	// Multiple-character operators

	TokenTypeLessThanOrEqual
	TokenTypeNotEquals
	TokenTypeGreaterThanOrEqual

	TokenTypeUnknown
)

func NewToken(tokenType TokenType, offset int, text string) Token {
	return Token{
		Type:   tokenType,
		Offset: offset,
		Text:   text,
	}
}
