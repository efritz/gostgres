package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/syntax/tokens"
)

func (p *parser) parseIdent() (string, error) {
	name, err := p.mustAdvance(isType(tokens.TokenTypeIdent))
	if err != nil {
		return "", err
	}

	return name.Text, nil
}

// parenthesized := `(` T `)`
func parseParenthesized[T any](p *parser, f func() (T, error)) (result T, _ error) {
	if _, err := p.mustAdvance(isType(tokens.TokenTypeLeftParen)); err != nil {
		return result, err
	}

	result, err := f()
	if err != nil {
		return result, err
	}

	if _, err := p.mustAdvance(isType(tokens.TokenTypeRightParen)); err != nil {
		return result, err
	}

	return result, nil
}

// commaSeparatedList := T [, ...]
func parseCommaSeparatedList[T any](p *parser, f func() (T, error)) (ts []T, _ error) {
	for {
		if t, err := f(); err != nil {
			return nil, err
		} else {
			ts = append(ts, t)
		}

		if !p.advanceIf(isType(tokens.TokenTypeComma)) {
			break
		}
	}

	return ts, nil
}

// parenthesizedCommaSeparatedList := [ `(` [ T [, ...] ] `)` ]
func parseParenthesizedCommaSeparatedList[T any](p *parser, optional bool, allowEmpty bool, f func() (T, error)) ([]T, error) {
	if next := p.peek(0); next.Type != tokens.TokenTypeLeftParen {
		if !optional {
			return nil, fmt.Errorf("expected parenthesized list (near %s)", next.Text)
		}

		return nil, nil
	}

	return parseParenthesized(p, func() ([]T, error) {
		if next := p.peek(0); next.Type == tokens.TokenTypeRightParen {
			if !allowEmpty {
				return nil, fmt.Errorf("unexpected zero-element list (near %s)", next.Text)
			}

			return nil, nil
		}

		return parseCommaSeparatedList(p, f)
	})
}
