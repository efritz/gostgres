package parsing

import (
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

// where := [ `WHERE` expression ]
func (p *parser) parseWhere() (impls.Expression, bool, error) {
	if !p.advanceIf(isType(tokens.TokenTypeWhere)) {
		return nil, false, nil
	}

	whereExpression, err := p.parseRootExpression()
	if err != nil {
		return nil, false, err
	}

	return whereExpression, true, nil
}

// returning := [`RETURNING` selectExpressions]
func (p *parser) parseReturning() (returningExpressions []projection.ProjectionExpression, err error) {
	if !p.advanceIf(isType(tokens.TokenTypeReturning)) {
		return nil, nil
	}

	return p.parseSelectExpressions()
}
