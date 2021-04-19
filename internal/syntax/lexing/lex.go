package lexing

import (
	"strings"

	"github.com/efritz/gostgres/internal/syntax/tokens"
)

func Lex(text string) (filteredTokens []tokens.Token) {
	lexer := newLexer(text)

loop:
	for {
		token := lexer.next()
		switch token.Type {
		case tokens.TokenTypeEOF:
			break loop
		case tokens.TokenTypeWhitespace:
			continue
		case tokens.TokenTypeIdent:
			if tokenType, ok := keywordSet[strings.ToLower(token.Text)]; ok {
				token.Type = tokenType
			}
		}

		filteredTokens = append(filteredTokens, token)
	}

	return filteredTokens
}
