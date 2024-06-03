package parsing

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/access"
	"github.com/efritz/gostgres/internal/execution/queries/aggregate"
	"github.com/efritz/gostgres/internal/execution/queries/alias"
	"github.com/efritz/gostgres/internal/execution/queries/combination"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/queries/joins"
	"github.com/efritz/gostgres/internal/execution/queries/limit"
	"github.com/efritz/gostgres/internal/execution/queries/mutation"
	"github.com/efritz/gostgres/internal/execution/queries/order"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/efritz/gostgres/internal/syntax/tokens"
)

type Builder interface {
	Build(ctx BuildContext) (queries.Node, error)
}

type BuildContext struct {
	Tables TableGetter
}

type TableGetter interface {
	Get(name string) (impls.Table, bool)
}

//
//

type SelectBuilder struct {
	SimpleSelect    *SimpleSelectDescription
	OrderExpression impls.OrderExpression
	Limit           *int
	Offset          *int
}

func (b *SelectBuilder) Build(ctx BuildContext) (queries.Node, error) {
	return b.TableExpression(ctx)
}

func (b SelectBuilder) TableExpression(ctx BuildContext) (queries.Node, error) {
	node, err := b.SimpleSelect.From.TableExpression(ctx)
	if err != nil {
		return nil, err
	}

	if b.SimpleSelect.WhereExpression != nil {
		node = filter.NewFilter(node, b.SimpleSelect.WhereExpression)
	}

	if b.SimpleSelect.Groupings != nil {
	selectLoop:
		for _, selectExpression := range b.SimpleSelect.SelectExpressions {
			expression, alias, ok := projection.UnwrapAlias(selectExpression)
			if !ok {
				return nil, fmt.Errorf("cannot unwrap alias %q", selectExpression)
			}

			if len(expressions.Fields(expression)) > 0 {
				for _, grouping := range b.SimpleSelect.Groupings {
					if grouping.Equal(expression) || grouping.Equal(expressions.NewNamed(fields.NewField("", alias, types.TypeAny))) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil,  fmt.Errorf("%q not in group by", expression)
			}
		}

		node = aggregate.NewHashAggregate(node, b.SimpleSelect.Groupings, b.SimpleSelect.SelectExpressions)
		b.SimpleSelect.SelectExpressions = nil
	}

	if len(b.SimpleSelect.Combinations) != 0 {
		if b.SimpleSelect.SelectExpressions != nil {
			newNode, err := projection.NewProjection(node, b.SimpleSelect.SelectExpressions)
			if err != nil {
				return nil, err
			}
			node = newNode
			b.SimpleSelect.SelectExpressions = nil
		}

		for _, c := range b.SimpleSelect.Combinations {
			var factory func(left, right queries.Node, distinct bool) (queries.Node, error)
			switch c.Type {
			case tokens.TokenTypeUnion:
				factory = combination.NewUnion
			case tokens.TokenTypeIntersect:
				factory = combination.NewIntersect
			case tokens.TokenTypeExcept:
				factory = combination.NewExcept
			}

			right, err := c.SimpleSelectDescription.Build(ctx)
			if err != nil {
				return nil, err
			}

			newNode, err := factory(node, right, c.Distinct)
			if err != nil {
				return nil, err
			}
			node = newNode
		}
	}

	if b.OrderExpression != nil {
		node = order.NewOrder(node, b.OrderExpression)
	}
	if b.Offset != nil {
		node = limit.NewOffset(node, *b.Offset)
	}
	if b.Limit != nil {
		node = limit.NewLimit(node, *b.Limit)
	}

	fmt.Printf("BUILDING SELECT\n")
	if b.SimpleSelect.SelectExpressions != nil {
		return projection.NewProjection(node, b.SimpleSelect.SelectExpressions)
	}
	return node, nil
}

//
//

type SimpleSelectDescription struct {
	SelectExpressions []projection.ProjectionExpression
	From              TableExpressionDescription
	WhereExpression   impls.Expression
	Groupings         []impls.Expression
	Combinations      []*CombinationDescription
}

type CombinationDescription struct {
	Type                    tokens.TokenType
	Distinct                bool
	SimpleSelectDescription BaseTableExpressionDescription
}

type ValuesBuilder struct {
	RowFields         []fields.Field
	AllRowExpressions [][]impls.Expression
}

func (b *ValuesBuilder) Build(ctx BuildContext) (queries.Node, error) {
	return b.TableExpression(ctx)
}

func (b ValuesBuilder) TableExpression(ctx BuildContext) (queries.Node, error) {
	fmt.Printf("BUILDING VALUES\n")
	return access.NewValues(b.RowFields, b.AllRowExpressions), nil
}

//
//

type TableExpressionDescription struct {
	Base  AliasedBaseTableExpressionDescription
	Joins []Join
}

type AliasedBaseTableExpressionDescription struct {
	BaseTableExpression BaseTableExpressionDescription
	Alias               *TableAlias
}

type BaseTableExpressionDescription interface {
	Builder
	TableExpression(ctx BuildContext) (queries.Node, error)
}

type TableReference struct {
	Name string
}

func (r TableReference) Build(ctx BuildContext) (queries.Node, error) {
	return r.TableExpression(ctx)
}

func (r TableReference) TableExpression(ctx BuildContext) (queries.Node, error) {
	table, ok := ctx.Tables.Get(r.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", r.Name)
	}

	return access.NewAccess(table), nil
}

func (e TableExpressionDescription) Build(ctx BuildContext) (queries.Node, error) {
	return e.TableExpression(ctx)
}

func (e TableExpressionDescription) TableExpression(ctx BuildContext) (queries.Node, error) {
	node, err := e.Base.BaseTableExpression.TableExpression(ctx)
	if err != nil {
		return nil, err
	}

	if e.Base.Alias != nil {
		aliasName := e.Base.Alias.TableAlias
		columnNames := e.Base.Alias.ColumnAliases

		node = alias.NewAlias(node, aliasName)

		if len(columnNames) > 0 {
			var fields []fields.Field
			for _, f := range node.Fields() {
				if !f.Internal() {
					fields = append(fields, f)
				}
			}

			if len(columnNames) != len(fields) {
				return nil, fmt.Errorf("wrong number of fields in alias")
			}

			projectionExpressions := make([]projection.ProjectionExpression, 0, len(fields))
			for i, field := range fields {
				projectionExpressions = append(projectionExpressions, projection.NewAliasProjectionExpression(expressions.NewNamed(field), columnNames[i]))
			}

			node, err = projection.NewProjection(node, projectionExpressions)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, j := range e.Joins {
		right, err := j.Table.TableExpression(ctx)
		if err != nil {
			return nil, err
		}

		node = joins.NewJoin(node, right, j.Condition)
	}

	return node, nil
}

type TableAlias struct {
	TableAlias    string
	ColumnAliases []string
}

type Join struct {
	Table     TableExpressionDescription
	Condition impls.Expression
}

//
//

type TableDescription struct {
	Name      string
	AliasName string
}

//
//

type InsertBuilder struct {
	TableDescription     TableDescription
	ColumnNames          []string
	Node                 BaseTableExpressionDescription
	ReturningExpressions []projection.ProjectionExpression
}

func (b *InsertBuilder) Build(ctx BuildContext) (queries.Node, error) {
	node, err := b.Node.Build(ctx)
	if err != nil {
		return nil, err
	}

	table, ok := ctx.Tables.Get(b.TableDescription.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.TableDescription.Name)
	}

	fmt.Printf("BUILDING INSERT\n")
	return mutation.NewInsert(node, table, b.TableDescription.Name, b.TableDescription.AliasName, b.ColumnNames, b.ReturningExpressions)
}

//
//

type UpdateBuilder struct {
	TableDescription     TableDescription
	SetExpressions       []SetExpression
	FromExpressions      []TableExpressionDescription
	WhereExpression      impls.Expression
	ReturningExpressions []projection.ProjectionExpression
}

type SetExpression struct {
	Name       string
	Expression impls.Expression
}

func (b *UpdateBuilder) Build(ctx BuildContext) (queries.Node, error) {
	table, ok := ctx.Tables.Get(b.TableDescription.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.TableDescription.Name)
	}

	node := access.NewAccess(table)
	if b.TableDescription.AliasName != "" {
		node = alias.NewAlias(node, b.TableDescription.AliasName)
	}

	if b.FromExpressions != nil {
		node = joinNodes2(ctx, node, b.FromExpressions)
	}

	if b.WhereExpression != nil {
		node = filter.NewFilter(node, b.WhereExpression)
	}

	relationName := b.TableDescription.Name
	if b.TableDescription.AliasName != "" {
		relationName = b.TableDescription.AliasName
	}
	tidField := fields.NewField(relationName, rows.TIDName, types.TypeBigInteger)

	node, err := projection.NewProjection(node, []projection.ProjectionExpression{
		projection.NewAliasProjectionExpression(expressions.NewNamed(tidField), rows.TIDName),
		projection.NewTableWildcardProjectionExpression(relationName),
	})
	if err != nil {
		return nil, err
	}

	node = alias.NewAlias(node, b.TableDescription.Name)

	setExpressions := make([]mutation.SetExpression, len(b.SetExpressions))
	for i, setExpression := range b.SetExpressions {
		setExpressions[i] = mutation.SetExpression{
			Name:       setExpression.Name,
			Expression: setExpression.Expression,
		}
	}

	fmt.Printf("BUILDING UPDATE\n")
	return mutation.NewUpdate(node, table, setExpressions, b.TableDescription.AliasName, b.ReturningExpressions)
}

//
//

type DeleteBuilder struct {
	TableDescription     TableDescription
	UsingExpressions     []TableExpressionDescription
	WhereExpression      impls.Expression
	returningExpressions []projection.ProjectionExpression
}

func (b *DeleteBuilder) Build(ctx BuildContext) (queries.Node, error) {
	table, ok := ctx.Tables.Get(b.TableDescription.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.TableDescription.Name)
	}

	node := access.NewAccess(table)
	if b.TableDescription.AliasName != "" {
		node = alias.NewAlias(node, b.TableDescription.AliasName)
	}
	if len(b.UsingExpressions) > 0 {
		node = joinNodes2(ctx, node, b.UsingExpressions)
	}
	if b.WhereExpression != nil {
		node = filter.NewFilter(node, b.WhereExpression)
	}

	relationName := b.TableDescription.Name
	if b.TableDescription.AliasName != "" {
		relationName = b.TableDescription.AliasName
	}
	tidField := fields.NewField(relationName, rows.TIDName, types.TypeBigInteger)

	node, err := projection.NewProjection(node, []projection.ProjectionExpression{
		projection.NewAliasProjectionExpression(expressions.NewNamed(tidField), rows.TIDName),
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("BUILDING DELETE\n")
	return mutation.NewDelete(node, table, b.TableDescription.AliasName, b.returningExpressions)
}

//
//

func joinNodes2(ctx BuildContext, left queries.Node, expressions []TableExpressionDescription) queries.Node {
	for _, expression := range expressions {
		right, err := expression.TableExpression(ctx)
		if err != nil {
			return nil
		}

		left = joins.NewJoin(left, right, nil)
	}

	return left
}
