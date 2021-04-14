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

func (l *lexer) scan(filter func(r rune) bool) (string, bool) {
	startOfToken := l.cursor
	for l.advanceIf(filter) {
		// no-op, advancing position
	}

	return l.text[startOfToken:l.cursor], startOfToken != l.cursor
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

var scanners = map[TokenType]func(r rune) bool{
	TokenTypeNumber:     isDigit,
	TokenTypeIdent:      isIdent,
	TokenTypeWhitespace: isSpace,
}

func isDigit(r rune) bool {
	return ('0' <= r && r <= '9')
}

func isIdent(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}
