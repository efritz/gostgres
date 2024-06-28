package parsing

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/ast"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type Query interface {
	Execute(ctx impls.Context, w protocol.ResponseWriter)
}

func Parse(tableGetter ast.TableGetter, tokenStream []tokens.Token) (Query, error) {
	parser := newParser(tokenStream)
	statement, err := parser.parseStatement(tableGetter)
	if err != nil {
		return nil, err
	}

	_ = parser.advanceIf(isType(tokens.TokenTypeSemicolon))
	if parser.cursor < len(parser.tokens) {
		return nil, fmt.Errorf("unexpected tokens at end of statement (near %s)", parser.tokens[parser.cursor].Text)
	}

	return statement, nil
}

func SplitStatements(input string) []string {
	var filtered []string
	for _, s := range strings.SplitAfter(stripComments(input), ";") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}

	return filtered
}

func stripComments(input string) string {
	var lines []string
	for _, line := range strings.Split(string(input), "\n") {
		lines = append(lines, strings.TrimRightFunc(strings.Split(line, "--")[0], unicode.IsSpace))
	}

	return strings.Join(lines, "\n")
}
