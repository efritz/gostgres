package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/explain"
	"github.com/efritz/gostgres/internal/queries/joins"
	"github.com/efritz/gostgres/internal/syntax/tokens"
	"github.com/efritz/gostgres/internal/table"
)

type parser struct {
	tokens             []tokens.Token
	cursor             int
	tables             TableGetter
	ddlParsers         map[tokens.TokenType]statementParserFunc
	explainableParsers map[tokens.TokenType]statementParserFunc
	prefixParsers      map[tokens.TokenType]prefixParserFunc
	infixParsers       map[tokens.TokenType]infixParserFunc
}

type TableGetter interface {
	GetTable(name string) (*table.Table, bool)
}

type tokenFilterFunc func(token tokens.Token) bool
type statementParserFunc func(token tokens.Token) (queries.Node, error)
type prefixParserFunc func(token tokens.Token) (expressions.Expression, error)
type infixParserFunc func(left expressions.Expression, token tokens.Token) (expressions.Expression, error)
type unaryExpressionParserFunc func(expression expressions.Expression) expressions.Expression
type binaryExpressionParserFunc func(left, right expressions.Expression) expressions.Expression
type ternaryExpressionParserFunc func(left, middle, right expressions.Expression) expressions.Expression

func newParser(tokenStream []tokens.Token, tables TableGetter) *parser {
	parser := &parser{
		tokens: tokenStream,
		tables: tables,
	}
	parser.ddlParsers = map[tokens.TokenType]statementParserFunc{
		tokens.TokenTypeCreate: parser.parseCreate,
	}
	parser.explainableParsers = map[tokens.TokenType]statementParserFunc{
		tokens.TokenTypeSelect: parser.parseSelect,
		tokens.TokenTypeInsert: parser.parseInsert,
		tokens.TokenTypeUpdate: parser.parseUpdate,
		tokens.TokenTypeDelete: parser.parseDelete,
	}
	parser.prefixParsers = map[tokens.TokenType]prefixParserFunc{
		tokens.TokenTypeIdent:     parser.parseNamedExpression,
		tokens.TokenTypeNumber:    parser.parseNumericLiteralExpression,
		tokens.TokenTypeString:    parser.parseStringLiteralExpression,
		tokens.TokenTypeFalse:     parser.parseBooleanLiteralExpression,
		tokens.TokenTypeNot:       parser.parseUnary(expressions.NewNot),
		tokens.TokenTypeNull:      parser.parseNullLiteralExpression,
		tokens.TokenTypeTrue:      parser.parseBooleanLiteralExpression,
		tokens.TokenTypeMinus:     parser.parseUnary(expressions.NewUnaryMinus),
		tokens.TokenTypeLeftParen: parser.parseParenthesizedExpression,
		tokens.TokenTypePlus:      parser.parseUnary(expressions.NewUnaryPlus),
	}
	parser.infixParsers = map[tokens.TokenType]infixParserFunc{
		tokens.TokenTypeAnd:                 parser.parseBinary(PrecedenceConditionalAnd, expressions.NewAnd),
		tokens.TokenTypeOr:                  parser.parseBinary(PrecedenceConditionalOr, expressions.NewOr),
		tokens.TokenTypeMinus:               parser.parseBinary(PrecedenceAdditive, expressions.NewSubtraction),
		tokens.TokenTypeAsterisk:            parser.parseBinary(PrecedenceMultiplicative, expressions.NewMultiplication),
		tokens.TokenTypeSlash:               parser.parseBinary(PrecedenceMultiplicative, expressions.NewDivision),
		tokens.TokenTypePlus:                parser.parseBinary(PrecedenceAdditive, expressions.NewAddition),
		tokens.TokenTypeLessThan:            parser.parseBinary(PrecedenceComparison, expressions.NewLessThan),
		tokens.TokenTypeEquals:              parser.parseBinary(PrecedenceEquality, expressions.NewEquals),
		tokens.TokenTypeGreaterThan:         parser.parseBinary(PrecedenceComparison, expressions.NewGreaterThan),
		tokens.TokenTypeLessThanOrEqual:     parser.parseBinary(PrecedenceComparison, expressions.NewLessThanEquals),
		tokens.TokenTypeNotEquals:           negate(parser.parseBinary(PrecedenceEquality, expressions.NewEquals)),
		tokens.TokenTypeGreaterThanOrEqual:  parser.parseBinary(PrecedenceComparison, expressions.NewGreaterThanEquals),
		tokens.TokenTypeIsTrue:              parser.parsePostfix(PrecedencePostfix, expressions.NewIsTrue),
		tokens.TokenTypeIsNotTrue:           negate(parser.parsePostfix(PrecedencePostfix, expressions.NewIsTrue)),
		tokens.TokenTypeIsFalse:             parser.parsePostfix(PrecedencePostfix, expressions.NewIsFalse),
		tokens.TokenTypeIsNotFalse:          negate(parser.parsePostfix(PrecedencePostfix, expressions.NewIsFalse)),
		tokens.TokenTypeIsNull:              parser.parsePostfix(PrecedencePostfix, expressions.NewIsNull),
		tokens.TokenTypeIsNotNull:           negate(parser.parsePostfix(PrecedencePostfix, expressions.NewIsNull)),
		tokens.TokenTypeIsUnknown:           parser.parsePostfix(PrecedencePostfix, expressions.NewIsUnknown),
		tokens.TokenTypeIsNotUnknown:        negate(parser.parsePostfix(PrecedencePostfix, expressions.NewIsUnknown)),
		tokens.TokenTypeConcat:              parser.parseBinary(PrecedenceGenericOperator, expressions.NewConcat),
		tokens.TokenTypeIsDistinctFrom:      parser.parseBinary(PrecedenceIs, expressions.NewIsDistinctFrom),
		tokens.TokenTypeIsNotDistinctFrom:   negate(parser.parseBinary(PrecedenceIs, expressions.NewIsDistinctFrom)),
		tokens.TokenTypeLike:                parser.parseBinary(PrecedenceLike, expressions.NewLike),
		tokens.TokenTypeNotLike:             negate(parser.parseBinary(PrecedenceLike, expressions.NewLike)),
		tokens.TokenTypeILike:               parser.parseBinary(PrecedenceLike, expressions.NewILike),
		tokens.TokenTypeNotILike:            negate(parser.parseBinary(PrecedenceLike, expressions.NewILike)),
		tokens.TokenTypeBetween:             parser.parseBetween(expressions.NewBetween),
		tokens.TokenTypeNotBetween:          negate(parser.parseBetween(expressions.NewBetween)),
		tokens.TokenTypeBetweenSymmetric:    parser.parseBetween(expressions.NewBetweenSymmetric),
		tokens.TokenTypeNotBetweenSymmetric: negate(parser.parseBetween(expressions.NewBetweenSymmetric)),
	}

	return parser
}

// statement := ddlStatement [`EXPLAIN`] explainableStatement
// ddlStatement := `CREATE` create
// explainableStatement := `SELECT` select | `INSERT` insert | `UPDATE` update | `DELETE` delete
func (p *parser) parseStatement() (queries.Node, error) {
	for tokenType, parser := range p.ddlParsers {
		token := p.current()
		if p.advanceIf(isType(tokenType)) {
			return parser(token)
		}
	}

	isExplain := false
	if p.advanceIf(isType(tokens.TokenTypeExplain)) {
		isExplain = true
	}

	for tokenType, parser := range p.explainableParsers {
		token := p.current()
		if p.advanceIf(isType(tokenType)) {
			node, err := parser(token)
			if err != nil {
				fmt.Printf("> %#v\n> %#v\n> %#v\n", token, node, err)
				return nil, err
			}

			if isExplain {
				node = explain.NewExplain(node)
			}

			return node, nil
		}
	}

	return nil, fmt.Errorf("expected start of statement (near %s)", p.current().Text)
}

func (p *parser) current() tokens.Token {
	return p.peek(0)
}

func (p *parser) peek(n int) tokens.Token {
	if p.cursor+n >= len(p.tokens) {
		return tokens.InvalidToken
	}

	return p.tokens[p.cursor+n]
}

func (p *parser) advance() tokens.Token {
	r := p.current()
	p.cursor++
	return r
}

func (p *parser) advanceIf(filters ...tokenFilterFunc) bool {
	start := p.cursor
	for _, filter := range filters {
		if !filter(p.current()) {
			p.cursor = start
			return false
		}

		p.cursor++
	}

	return true
}

func (p *parser) mustAdvance(filter tokenFilterFunc) (tokens.Token, error) {
	current := p.advance()
	if !filter(current) {
		return tokens.InvalidToken, fmt.Errorf("unexpected token (near %s)", current.Text)
	}

	return current, nil
}

func isType(tokenType tokens.TokenType) tokenFilterFunc {
	return func(t tokens.Token) bool {
		return t.Type == tokenType
	}
}

func negate(parserFunc infixParserFunc) infixParserFunc {
	return func(left expressions.Expression, token tokens.Token) (expressions.Expression, error) {
		expression, err := parserFunc(left, token)
		if err != nil {
			return nil, err
		}

		return expressions.NewNot(expression), nil
	}
}

func joinNodes(expressions []queries.Node) queries.Node {
	if len(expressions) == 0 {
		return nil
	}

	left := expressions[0]
	for _, right := range expressions[1:] {
		left = joins.NewJoin(left, right, nil)
	}

	return left
}
