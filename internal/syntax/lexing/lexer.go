package lexing

import "github.com/efritz/gostgres/internal/syntax/tokens"

type lexer struct {
	text   string
	cursor int
}

func newLexer(text string) *lexer {
	return &lexer{
		text: text,
	}
}

var keywordSet = map[string]tokens.TokenType{
	"add":        tokens.TokenTypeAdd,
	"all":        tokens.TokenTypeAll,
	"alter":      tokens.TokenTypeAlter,
	"and":        tokens.TokenTypeAnd,
	"as":         tokens.TokenTypeAs,
	"asc":        tokens.TokenTypeAscending,
	"between":    tokens.TokenTypeBetween,
	"by":         tokens.TokenTypeBy,
	"check":      tokens.TokenTypeCheck,
	"constraint": tokens.TokenTypeConstraint,
	"create":     tokens.TokenTypeCreate,
	"default":    tokens.TokenTypeDefault,
	"delete":     tokens.TokenTypeDelete,
	"desc":       tokens.TokenTypeDescending,
	"distinct":   tokens.TokenTypeDistinct,
	"except":     tokens.TokenTypeExcept,
	"explain":    tokens.TokenTypeExplain,
	"false":      tokens.TokenTypeFalse,
	"foreign":    tokens.TokenTypeForeign,
	"from":       tokens.TokenTypeFrom,
	"group":      tokens.TokenTypeGroup,
	"ilike":      tokens.TokenTypeILike,
	"index":      tokens.TokenTypeIndex,
	"insert":     tokens.TokenTypeInsert,
	"intersect":  tokens.TokenTypeIntersect,
	"into":       tokens.TokenTypeInto,
	"is":         tokens.TokenTypeIs,
	"isnull":     tokens.TokenTypeIsNull,
	"join":       tokens.TokenTypeJoin,
	"key":        tokens.TokenTypeKey,
	"like":       tokens.TokenTypeLike,
	"limit":      tokens.TokenTypeLimit,
	"not":        tokens.TokenTypeNot,
	"notnull":    tokens.TokenTypeIsNotNull,
	"null":       tokens.TokenTypeNull,
	"offset":     tokens.TokenTypeOffset,
	"on":         tokens.TokenTypeOn,
	"or":         tokens.TokenTypeOr,
	"order":      tokens.TokenTypeOrder,
	"primary":    tokens.TokenTypePrimary,
	"references": tokens.TokenTypeReferences,
	"returning":  tokens.TokenTypeReturning,
	"select":     tokens.TokenTypeSelect,
	"sequence":   tokens.TokenTypeSequence,
	"set":        tokens.TokenTypeSet,
	"symmetric":  tokens.TokenTypeSymmetric,
	"table":      tokens.TokenTypeTable,
	"true":       tokens.TokenTypeTrue,
	"union":      tokens.TokenTypeUnion,
	"unique":     tokens.TokenTypeUnique,
	"unknown":    tokens.TokenTypeKwUnknown,
	"update":     tokens.TokenTypeUpdate,
	"using":      tokens.TokenTypeUsing,
	"values":     tokens.TokenTypeValues,
	"where":      tokens.TokenTypeWhere,
}

var punctuationMap = map[rune]tokens.TokenType{
	0:   tokens.TokenTypeEOF,
	'-': tokens.TokenTypeMinus,
	',': tokens.TokenTypeComma,
	';': tokens.TokenTypeSemicolon,
	'.': tokens.TokenTypeDot,
	'(': tokens.TokenTypeLeftParen,
	')': tokens.TokenTypeRightParen,
	'*': tokens.TokenTypeAsterisk,
	'/': tokens.TokenTypeSlash,
	'+': tokens.TokenTypePlus,
	'<': tokens.TokenTypeLessThan,
	'=': tokens.TokenTypeEquals,
	'>': tokens.TokenTypeGreaterThan,
}

var multipleCharacterPunctuationMap = map[rune]map[string]tokens.TokenType{
	'!': {"=": tokens.TokenTypeNotEquals},
	'<': {"=": tokens.TokenTypeLessThanOrEqual, ">": tokens.TokenTypeNotEquals},
	'>': {"=": tokens.TokenTypeGreaterThanOrEqual},
	'|': {"|": tokens.TokenTypeConcat},
}

func (l *lexer) next() tokens.Token {
	startOfToken := l.cursor

	for tokenType, filter := range scanners {
		if value, ok := l.scan(filter); ok {
			return tokens.NewToken(tokenType, startOfToken, value)
		}
	}

	r := l.advance()

	punctuationScanners := []func(r rune) (tokens.TokenType, string, bool){
		l.scanMultipleCharacterPunctuation,
		l.scanPunctuation,
	}

	for _, scan := range punctuationScanners {
		if tokenType, value, ok := scan(r); ok {
			return tokens.NewToken(tokenType, startOfToken, value)
		}
	}

	panic("unreachable")
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

func (l *lexer) scanMultipleCharacterPunctuation(r rune) (tokens.TokenType, string, bool) {
	suffixMap, ok := multipleCharacterPunctuationMap[r]
	if ok {
		for suffix, tokenType := range suffixMap {
			if l.peekSubstring(len(suffix)) == suffix {
				l.cursor += len(suffix)
				return tokenType, string(r) + suffix, true
			}
		}
	}

	return tokens.TokenTypeInvalid, "", false
}

func (l *lexer) scanPunctuation(r rune) (tokens.TokenType, string, bool) {
	tokenType, ok := punctuationMap[r]
	if !ok {
		tokenType = tokens.TokenTypeInvalid
	}

	return tokenType, string(r), true
}

func (l *lexer) current() rune {
	if l.cursor >= len(l.text) {
		return 0
	}

	return rune(l.text[l.cursor])
}

func (l *lexer) peekSubstring(dist int) string {
	end := l.cursor + dist
	if end >= len(l.text) {
		end = len(l.text)
	}

	return l.text[l.cursor:end]
}

func (l *lexer) advance() rune {
	r := l.current()
	l.cursor++
	return r
}

func (l *lexer) advanceIf(filter func(r rune) bool) bool {
	if !filter(l.current()) {
		return false
	}

	l.cursor++
	return true
}
