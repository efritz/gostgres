package lexing

import "github.com/efritz/gostgres/internal/syntax/tokens"

type scanner struct {
	canStart    func(r rune) bool
	canContinue func(r rune) bool
	inclusive   bool
}

var scanners = map[tokens.TokenType]scanner{
	tokens.TokenTypeWhitespace: {isSpace, isSpace, true},
	tokens.TokenTypeString:     {isQuote, isNotQuote, false},
	tokens.TokenTypeIdent:      {isIdent, isIdentOrDigit, true},
	tokens.TokenTypeNumber:     {isDigit, isDigit, true},
}

func isSpace(r rune) bool        { return r == ' ' || r == '\t' || r == '\n' }
func isQuote(r rune) bool        { return r == '\'' }
func isNotQuote(r rune) bool     { return r != '\'' }
func isIdent(r rune) bool        { return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || r == '_' }
func isDigit(r rune) bool        { return ('0' <= r && r <= '9') }
func isIdentOrDigit(r rune) bool { return isIdent(r) || isDigit(r) }
