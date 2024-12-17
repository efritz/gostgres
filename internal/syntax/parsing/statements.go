package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

func (p *parser) initDDLParsers() {
	p.ddlParsers = ddlParsers{
		tokens.TokenTypeCreate: p.parseCreate,
		tokens.TokenTypeAlter:  p.parseAlter,
	}
}

func (p *parser) initStatementParsers() {
	p.queryParsers = explainableParsers{
		tokens.TokenTypeAnalyze: p.parseAnalyze,
		tokens.TokenTypeSelect:  p.parseSelectBuilder,
		tokens.TokenTypeInsert:  p.parseInsert,
		tokens.TokenTypeUpdate:  p.parseUpdate,
		tokens.TokenTypeDelete:  p.parseDelete,
	}
}

// statement := ddlStatement | analyzeStatement ( [ `EXPLAIN` ] explainableStatement )
// ddlStatement := ( `CREATE` createTail ) | ( `ALTER` alterTail )
// analyzeStatement := `ANALYZE` analyzeTail
// queryStatement := ( `SELECT` selectTail ) | ( `INSERT` insertTail ) | ( `UPDATE` updateTail ) | ( `DELETE` deleteTail )
func (p *parser) parseStatement(catalog impls.CatalogSet) (Query, error) {
	for tokenType, parser := range p.ddlParsers {
		token := p.current()
		if p.advanceIf(isType(tokenType)) {
			return parser(token)
		}
	}

	isExplain := false
	if p.advanceIf(isType(tokens.TokenTypeExplain)) {
		isExplain = true

		if p.peek(0).Type == tokens.TokenTypeAnalyze {
			return nil, fmt.Errorf("cannot explain analyze")
		}
	}

	for tokenType, parser := range p.queryParsers {
		token := p.current()
		if p.advanceIf(isType(tokenType)) {
			builder, err := parser(token)
			if err != nil {
				return nil, err
			}

			if err := builder.Resolve(impls.NewNodeResolutionContext(catalog)); err != nil {
				return nil, err
			}

			node, err := builder.Build()
			if err != nil {
				return nil, err
			}

			if isExplain {
				node = plan.NewExplain(node)
			}

			return plan.NewQuery(node), nil
		}
	}

	return nil, fmt.Errorf("expected start of statement (near %s)", p.current().Text)
}
