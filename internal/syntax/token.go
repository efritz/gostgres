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

	// Keywords

	TokenTypeAnd
	TokenTypeAs
	TokenTypeBy
	TokenTypeFalse
	TokenTypeFrom
	TokenTypeIs // TODO - IS NULL <-> IS NULL; IS NOT NULL <-> NOT NULL
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
	TokenTypeSelect
	TokenTypeTrue
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
