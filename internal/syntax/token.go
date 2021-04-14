package syntax

type Token struct {
	Type   TokenType
	Offset int
	Text   string
}

type TokenType int

const (
	TokenTypeInvalid TokenType = iota
	TokenTypeEOF
	TokenTypeWhitespace
	TokenTypeKeyword
	TokenTypeIdent
	TokenTypeNumber
	TokenTypeComma
	TokenTypeSemicolon
	TokenTypeDot
	TokenTypeLeftParen
	TokenTypeRightParen
	TokenTypeAsterisk
	TokenTypeUnknown
)

func NewToken(tokenType TokenType, offset int, text string) Token {
	return Token{
		Type:   tokenType,
		Offset: offset,
		Text:   text,
	}
}

var InvalidToken = Token{
	Type: TokenTypeInvalid,
}
