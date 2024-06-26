package lexing

import (
	"sort"
	"strings"

	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type tokenSequenceReplacement struct {
	replacement tokens.TokenType
	pattern     []tokens.TokenType
}

var tokenSequenceReplacements = []tokenSequenceReplacement{
	{tokens.TokenTypeBetweenSymmetric, []tokens.TokenType{tokens.TokenTypeBetween, tokens.TokenTypeSymmetric}},
	{tokens.TokenTypeForeignKey, []tokens.TokenType{tokens.TokenTypeForeign, tokens.TokenTypeKey}},
	{tokens.TokenTypeIsDistinctFrom, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeDistinct, tokens.TokenTypeFrom}},
	{tokens.TokenTypeIsFalse, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeFalse}},
	{tokens.TokenTypeIsNotDistinctFrom, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeNot, tokens.TokenTypeDistinct, tokens.TokenTypeFrom}},
	{tokens.TokenTypeIsNotFalse, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeNot, tokens.TokenTypeFalse}},
	{tokens.TokenTypeIsNotNull, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeNot, tokens.TokenTypeNull}},
	{tokens.TokenTypeIsNotTrue, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeNot, tokens.TokenTypeTrue}},
	{tokens.TokenTypeIsNotUnknown, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeNot, tokens.TokenTypeKwUnknown}},
	{tokens.TokenTypeIsNull, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeNull}},
	{tokens.TokenTypeIsTrue, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeTrue}},
	{tokens.TokenTypeIsUnknown, []tokens.TokenType{tokens.TokenTypeIs, tokens.TokenTypeKwUnknown}},
	{tokens.TokenTypeNotBetween, []tokens.TokenType{tokens.TokenTypeNot, tokens.TokenTypeBetween}},
	{tokens.TokenTypeNotBetweenSymmetric, []tokens.TokenType{tokens.TokenTypeNot, tokens.TokenTypeBetween, tokens.TokenTypeSymmetric}},
	{tokens.TokenTypeNotILike, []tokens.TokenType{tokens.TokenTypeNot, tokens.TokenTypeILike}},
	{tokens.TokenTypeNotLike, []tokens.TokenType{tokens.TokenTypeNot, tokens.TokenTypeLike}},
	{tokens.TokenTypeNotNull, []tokens.TokenType{tokens.TokenTypeNot, tokens.TokenTypeNull}},
	{tokens.TokenTypePrimaryKey, []tokens.TokenType{tokens.TokenTypePrimary, tokens.TokenTypeKey}},
}

func init() {
	sort.Slice(tokenSequenceReplacements, func(i, j int) bool {
		return len(tokenSequenceReplacements[j].pattern) < len(tokenSequenceReplacements[i].pattern)
	})
}

func Lex(text string) (filteredTokens []tokens.Token) {
	lexer := newLexer(text)

	bufferSize := 0
	for _, re := range tokenSequenceReplacements {
		if bufferSize < len(re.pattern) {
			bufferSize = len(re.pattern)
		}
	}

	var buffer []tokens.Token
	for {
		for len(buffer) < bufferSize {
			if token := lexer.next(); token.Type != tokens.TokenTypeWhitespace {
				buffer = append(buffer, hydrateKeywords(token))
			}
		}

		token, hasToken, newBuffer, ok := processBuffer(buffer)
		if !ok {
			break
		}
		if hasToken {
			filteredTokens = append(filteredTokens, token)
		}
		buffer = newBuffer
	}

	return filteredTokens
}

func hydrateKeywords(token tokens.Token) tokens.Token {
	if token.Type == tokens.TokenTypeIdent {
		if tokenType, ok := keywordSet[strings.ToLower(token.Text)]; ok {
			token.Type = tokenType
		}
	}

	return token
}

func processBuffer(buffer []tokens.Token) (tokens.Token, bool, []tokens.Token, bool) {
	if buffer[0].Type == tokens.TokenTypeEOF {
		return tokens.Token{}, false, nil, false
	}

	var bufferTypes []tokens.TokenType
	for _, t := range buffer {
		bufferTypes = append(bufferTypes, t.Type)
	}

	for _, re := range tokenSequenceReplacements {
		matches := true
		for i, t := range re.pattern {
			if bufferTypes[i] != t {
				matches = false
			}
		}
		if !matches {
			continue
		}

		texts := make([]string, 0, len(re.pattern))
		for i := range re.pattern {
			texts = append(texts, buffer[i].Text)
		}

		token := tokens.Token{
			Type:   re.replacement,
			Offset: buffer[0].Offset,
			Text:   strings.Join(texts, " "),
		}

		return token, true, buffer[len(re.pattern):], true
	}

	return buffer[0], true, buffer[1:], true
}
