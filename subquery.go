package liqu

import (
	"fmt"
	"strings"
)

type (
	SubQuery struct {
		model       string
		field       string
		from        string
		joins       []string
		fieldLocal  string
		fieldParent string
		query       *query
		conditions  *ConditionBuilder
	}
)

func (l *Liqu) WithSubQuery(query *SubQuery) *Liqu {
	l.subQueries = append(l.subQueries, query)

	return l
}

func (l *Liqu) parseSubQueries() error {
	for _, sq := range l.subQueries {
		if _, ok := l.registry[sq.model].fieldDatabase[sq.field]; !ok {
			return fmt.Errorf("invalid subquery for model %s with field %s", sq.model, sq.field)
		}

		l.registry[sq.model].branch.subQuery[sq.field] = sq
	}

	return nil
}

func NewSubQuery(model string, field string) *SubQuery {
	return &SubQuery{
		model:      model,
		field:      field,
		query:      newBaseQuery(),
		conditions: NewConditionBuilder(),
	}
}

func (sq *SubQuery) And(and string) *SubQuery {
	sq.conditions.AndRaw(and)

	return sq
}
func (sq *SubQuery) Relate(local string, parent string) *SubQuery {
	sq.fieldLocal = local
	sq.fieldParent = parent

	return sq
}

func (sq *SubQuery) Select(sel string) *SubQuery {
	sq.query.setSelect(sel)
	return sq
}

func (sq *SubQuery) SelectAggregate(funcName, field, alias string) *SubQuery {
	sq.query.setSelect(fmt.Sprintf("%s(%s) AS %s", funcName, field, alias))
	return sq
}

func (sq *SubQuery) From(from string) *SubQuery {
	sq.from = from
	sq.query.setFrom(from)
	return sq
}

func (sq *SubQuery) Join(joinTable, condition, joinType string) *SubQuery {
	join := fmt.Sprintf("%s JOIN %s ON %s", strings.ToUpper(joinType), joinTable, condition)
	sq.joins = append(sq.joins, join)
	return sq
}

func (sq *SubQuery) As(as string) *SubQuery {
	sq.from = as
	sq.query.setAs(as)
	return sq
}

func (sq *SubQuery) GroupBy(gb string) *SubQuery {
	sq.query.setGroupBy(gb)
	return sq
}

func (sq *SubQuery) OrderBy(ob string) *SubQuery {
	sq.query.setOrderBy(ob)
	return sq
}

func (sq *SubQuery) Limit(limit int) *SubQuery {
	sq.query.setLimit(&Filters{
		Page:    1,
		PerPage: limit,
	})
	return sq
}

func (sq *SubQuery) Build() string {
	sq.query.setWhere(sq.conditions.Build())

	if len(sq.joins) > 0 {
		sq.query.setJoin(strings.Join(sq.joins, " "))
	}

	return sq.query.Scrub()
}
