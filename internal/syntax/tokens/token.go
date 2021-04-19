package tokens

type Token struct {
	Type   TokenType
	Offset int
	Text   string
}

var InvalidToken = Token{
	Type: TokenTypeInvalid,
}

func NewToken(tokenType TokenType, offset int, text string) Token {
	return Token{
		Type:   tokenType,
		Offset: offset,
		Text:   text,
	}
}
