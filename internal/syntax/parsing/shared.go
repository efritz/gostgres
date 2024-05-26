package parsing

import (
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/syntax/tokens"
	"github.com/efritz/gostgres/internal/types"
)

// where := [ `WHERE` expression ]
func (p *parser) parseWhere() (types.Expression, bool, error) {
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
func (p *parser) parseReturning(name string) (returningExpressions []projection.ProjectionExpression, err error) {
	if !p.advanceIf(isType(tokens.TokenTypeReturning)) {
		return nil, nil
	}

	returningExpressions, err = p.parseSelectExpressions()
	if err != nil {
		return nil, err
	}
	if returningExpressions != nil {
		return returningExpressions, nil
	}

	return []projection.ProjectionExpression{projection.NewTableWildcardProjectionExpression(name)}, nil
}
