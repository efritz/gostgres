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
	Build() (queries.Node, error)
}

//
//

type SelectBuilder struct {
	simpleSelect    *SimpleSelectDescription
	orderExpression impls.OrderExpression
	limit           *int
	offset          *int
}

func (b *SelectBuilder) Build() (queries.Node, error) {
	node, err := b.simpleSelect.from.TableExpression()
	if err != nil {
		return nil, err
	}

	if b.simpleSelect.whereExpression != nil {
		node = filter.NewFilter(node, b.simpleSelect.whereExpression)
	}

	if b.simpleSelect.groupings != nil {
	selectLoop:
		for _, selectExpression := range b.simpleSelect.selectExpressions {
			expression, alias, ok := projection.UnwrapAlias(selectExpression)
			if !ok {
				return nil, fmt.Errorf("cannot unwrap alias %q", selectExpression)
			}

			if len(expressions.Fields(expression)) > 0 {
				for _, grouping := range b.simpleSelect.groupings {
					if grouping.Equal(expression) || grouping.Equal(expressions.NewNamed(fields.NewField("", alias, types.TypeAny))) {
						continue selectLoop
					}
				}

				// TODO - more lenient validation
				// return nil,  fmt.Errorf("%q not in group by", expression)
			}
		}

		node = aggregate.NewHashAggregate(node, b.simpleSelect.groupings, b.simpleSelect.selectExpressions)
		b.simpleSelect.selectExpressions = nil
	}

	if len(b.simpleSelect.combinations) != 0 {
		if b.simpleSelect.selectExpressions != nil {
			newNode, err := projection.NewProjection(node, b.simpleSelect.selectExpressions)
			if err != nil {
				return nil, err
			}
			node = newNode
			b.simpleSelect.selectExpressions = nil
		}

		for _, c := range b.simpleSelect.combinations {
			var factory func(left, right queries.Node, distinct bool) (queries.Node, error)
			switch c.Type {
			case tokens.TokenTypeUnion:
				factory = combination.NewUnion
			case tokens.TokenTypeIntersect:
				factory = combination.NewIntersect
			case tokens.TokenTypeExcept:
				factory = combination.NewExcept
			}

			right, err := c.SimpleSelectDescription.Build()
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

	if b.orderExpression != nil {
		node = order.NewOrder(node, b.orderExpression)
	}
	if b.offset != nil {
		node = limit.NewOffset(node, *b.offset)
	}
	if b.limit != nil {
		node = limit.NewLimit(node, *b.limit)
	}

	fmt.Printf("BUILDING SELECT\n")
	if b.simpleSelect.selectExpressions != nil {
		return projection.NewProjection(node, b.simpleSelect.selectExpressions)
	}
	return node, nil
}

//
//

type SimpleSelectDescription struct {
	selectExpressions []projection.ProjectionExpression
	from              TableExpressionDescription
	whereExpression   impls.Expression
	groupings         []impls.Expression
	combinations      []*CombinationDescription
}

type CombinationDescription struct {
	Type                    tokens.TokenType
	Distinct                bool
	SimpleSelectDescription SelectOrValuesBuilder
}

type ValuesBuilder struct {
	rowFields         []fields.Field
	allRowExpressions [][]impls.Expression
}

func (b *ValuesBuilder) Build() (queries.Node, error) {
	fmt.Printf("BUILDING VALUES\n")
	return access.NewValues(b.rowFields, b.allRowExpressions), nil
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
	TableExpression() (queries.Node, error)
}

type TableReference struct {
	tables TableGetter
	Name   string
}

func (r TableReference) TableExpression() (queries.Node, error) {
	table, ok := r.tables.Get(r.Name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", r.Name)
	}

	return access.NewAccess(table), nil
}

func (e TableExpressionDescription) TableExpression() (queries.Node, error) {
	node, err := e.Base.BaseTableExpression.TableExpression()
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
		right, err := j.table.TableExpression()
		if err != nil {
			return nil, err
		}

		node = joins.NewJoin(node, right, j.condition)
	}

	return node, nil
}

type TableAlias struct {
	TableAlias    string
	ColumnAliases []string
}

type Join struct {
	table     TableExpressionDescription
	condition impls.Expression
}

//
//

type SelectOrValuesBuilder interface {
	Builder
	BaseTableExpressionDescription
	selectOrValues()
}

func (SelectBuilder) selectOrValues() {}
func (ValuesBuilder) selectOrValues() {}

func (b SelectBuilder) TableExpression() (queries.Node, error) {
	return b.Build()
}

func (b ValuesBuilder) TableExpression() (queries.Node, error) {
	return b.Build()
}

//
//

type TableDescription struct {
	tables    TableGetter
	name      string
	aliasName string
}

//
//

type InsertBuilder struct {
	tableDescription     TableDescription
	columnNames          []string
	node                 SelectOrValuesBuilder
	returningExpressions []projection.ProjectionExpression
}

func (b *InsertBuilder) Build() (queries.Node, error) {
	node, err := b.node.Build()
	if err != nil {
		return nil, err
	}

	table, ok := b.tableDescription.tables.Get(b.tableDescription.name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.tableDescription.name)
	}

	fmt.Printf("BUILDING INSERT\n")
	return mutation.NewInsert(node, table, b.tableDescription.name, b.tableDescription.aliasName, b.columnNames, b.returningExpressions)
}

//
//

type UpdateBuilder struct {
	tableDescription     TableDescription
	setExpressions       []SetExpression
	fromExpressions      []TableExpressionDescription
	whereExpression      impls.Expression
	returningExpressions []projection.ProjectionExpression
}

type SetExpression struct {
	Name       string
	Expression impls.Expression
}

func (b *UpdateBuilder) Build() (queries.Node, error) {
	table, ok := b.tableDescription.tables.Get(b.tableDescription.name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.tableDescription.name)
	}

	node := access.NewAccess(table)
	if b.tableDescription.aliasName != "" {
		node = alias.NewAlias(node, b.tableDescription.aliasName)
	}

	if b.fromExpressions != nil {
		node = joinNodes2(node, b.fromExpressions)
	}

	if b.whereExpression != nil {
		node = filter.NewFilter(node, b.whereExpression)
	}

	relationName := b.tableDescription.name
	if b.tableDescription.aliasName != "" {
		relationName = b.tableDescription.aliasName
	}
	tidField := fields.NewField(relationName, rows.TIDName, types.TypeBigInteger)

	node, err := projection.NewProjection(node, []projection.ProjectionExpression{
		projection.NewAliasProjectionExpression(expressions.NewNamed(tidField), rows.TIDName),
		projection.NewTableWildcardProjectionExpression(relationName),
	})
	if err != nil {
		return nil, err
	}

	node = alias.NewAlias(node, b.tableDescription.name)

	setExpressions := make([]mutation.SetExpression, len(b.setExpressions))
	for i, setExpression := range b.setExpressions {
		setExpressions[i] = mutation.SetExpression{
			Name:       setExpression.Name,
			Expression: setExpression.Expression,
		}
	}

	fmt.Printf("BUILDING UPDATE\n")
	return mutation.NewUpdate(node, table, setExpressions, b.tableDescription.aliasName, b.returningExpressions)
}

//
//

type DeleteBuilder struct {
	tableDescription     TableDescription
	usingExpressions     []TableExpressionDescription
	whereExpression      impls.Expression
	returningExpressions []projection.ProjectionExpression
}

func (b *DeleteBuilder) Build() (queries.Node, error) {
	table, ok := b.tableDescription.tables.Get(b.tableDescription.name)
	if !ok {
		return nil, fmt.Errorf("unknown table %s", b.tableDescription.name)
	}

	node := access.NewAccess(table)
	if b.tableDescription.aliasName != "" {
		node = alias.NewAlias(node, b.tableDescription.aliasName)
	}
	if len(b.usingExpressions) > 0 {
		node = joinNodes2(node, b.usingExpressions)
	}
	if b.whereExpression != nil {
		node = filter.NewFilter(node, b.whereExpression)
	}

	relationName := b.tableDescription.name
	if b.tableDescription.aliasName != "" {
		relationName = b.tableDescription.aliasName
	}
	tidField := fields.NewField(relationName, rows.TIDName, types.TypeBigInteger)

	node, err := projection.NewProjection(node, []projection.ProjectionExpression{
		projection.NewAliasProjectionExpression(expressions.NewNamed(tidField), rows.TIDName),
	})
	if err != nil {
		return nil, err
	}

	fmt.Printf("BUILDING DELETE\n")
	return mutation.NewDelete(node, table, b.tableDescription.aliasName, b.returningExpressions)
}

//
//

func joinNodes2(left queries.Node, expressions []TableExpressionDescription) queries.Node {
	for _, expression := range expressions {
		right, err := expression.TableExpression()
		if err != nil {
			return nil
		}

		left = joins.NewJoin(left, right, nil)
	}

	return left
}
