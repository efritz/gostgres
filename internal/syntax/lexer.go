package syntax

import (
	"strings"
)

var keywordSet = map[string]TokenType{
	"and":     TokenTypeAnd,
	"as":      TokenTypeAs,
	"by":      TokenTypeBy,
	"false":   TokenTypeFalse,
	"from":    TokenTypeFrom,
	"insert":  TokenTypeInsert,
	"into":    TokenTypeInto,
	"is":      TokenTypeIs,
	"isnull":  TokenTypeIsNull,
	"join":    TokenTypeJoin,
	"limit":   TokenTypeLimit,
	"not":     TokenTypeNot,
	"notnull": TokenTypeNotNull,
	"null":    TokenTypeNull,
	"offset":  TokenTypeOffset,
	"on":      TokenTypeOn,
	"or":      TokenTypeOr,
	"order":   TokenTypeOrder,
	"select":  TokenTypeSelect,
	"true":    TokenTypeTrue,
	"values":  TokenTypeValues,
	"where":   TokenTypeWhere,
}

var punctuationMap = map[rune]TokenType{
	0:   TokenTypeEOF,
	'-': TokenTypeMinus,
	',': TokenTypeComma,
	';': TokenTypeSemicolon,
	'.': TokenTypeDot,
	'(': TokenTypeLeftParen,
	')': TokenTypeRightParen,
	'*': TokenTypeAsterisk,
	'/': TokenTypeSlash,
	'+': TokenTypePlus,
	'<': TokenTypeLessThan,
	'=': TokenTypeEquals,
	'>': TokenTypeGreaterThan,
}

var multipleCharacterPunctuationMap = map[rune]map[string]TokenType{
	'!': {"=": TokenTypeNotEquals},
	'<': {"=": TokenTypeLessThanOrEqual, ">": TokenTypeNotEquals},
	'>': {"=": TokenTypeGreaterThanOrEqual},
}

func Lex(text string) (tokens []Token) {
	lexer := lexer{text: text}
	for {
		token := lexer.next()
		if token.Type == TokenTypeEOF {
			break
		}
		if token.Type == TokenTypeWhitespace {
			continue
		}
		if token.Type == TokenTypeIdent {
			if tokenType, ok := keywordSet[strings.ToLower(token.Text)]; ok {
				token.Type = tokenType
			}
		}

		tokens = append(tokens, token)
	}

	return tokens
}

type lexer struct {
	text   string
	cursor int
}

func (l *lexer) next() Token {
	startOfToken := l.cursor

	for tokenType, filter := range scanners {
		if value, ok := l.scan(filter); ok {
			return NewToken(tokenType, startOfToken, value)
		}
	}

	r := l.advance()

	suffixMap, ok := multipleCharacterPunctuationMap[r]
	if ok {
		for suffix, tokenType := range suffixMap {
			if l.peek(len(suffix)) == suffix {
				l.cursor += len(suffix)
				return NewToken(tokenType, startOfToken, string(r)+suffix)
			}
		}
	}

	tokenType, ok := punctuationMap[r]
	if !ok {
		tokenType = TokenTypeInvalid
	}
	return NewToken(tokenType, startOfToken, string(r))
}

func (l *lexer) scan(filter scanner) (string, bool) {
	start := l.cursor

	if l.advanceIf(filter.canStart) {
		for l.advanceIf(filter.canContinue) {
			// no-op, advancing position
		}

		end := l.cursor
		if !filter.inclusive {
			start++
			l.cursor++
		}

		return l.text[start:end], true
	}

	return "", false
}

func (l *lexer) current() rune {
	if l.cursor >= len(l.text) {
		return 0
	}

	return rune(l.text[l.cursor])
}

func (l *lexer) advance() rune {
	r := l.current()
	l.cursor++
	return r
}

func (l *lexer) peek(dist int) string {
	end := l.cursor + dist
	if end >= len(l.text) {
		end = len(l.text)
	}

	return l.text[l.cursor:end]
}

func (l *lexer) advanceIf(filter func(r rune) bool) bool {
	if !filter(l.current()) {
		return false
	}

	l.cursor++
	return true
}

type scanner struct {
	canStart    func(r rune) bool
	canContinue func(r rune) bool
	inclusive   bool
}

var scanners = map[TokenType]scanner{
	TokenTypeWhitespace: {isSpace, isSpace, true},
	TokenTypeString:     {isQuote, isNotQuote, false},
	TokenTypeIdent:      {isIdent, isIdentOrDigit, true},
	TokenTypeNumber:     {isDigit, isDigit, true},
}

func isSpace(r rune) bool        { return r == ' ' || r == '\t' || r == '\n' }
func isQuote(r rune) bool        { return r == '\'' }
func isNotQuote(r rune) bool     { return r != '\'' }
func isIdent(r rune) bool        { return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || r == '_' }
func isDigit(r rune) bool        { return ('0' <= r && r <= '9') }
func isIdentOrDigit(r rune) bool { return isIdent(r) || isDigit(r) }
