package liqu

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	rootQuery         = "SELECT :totalRows: :select: FROM ( :from: :where: :groupBy: :orderBy:) :as: :join: :whereNulls: :limit:"
	baseQuery         = `SELECT :select: FROM ":from:" :as: :join: :where: :groupBy: :orderBy: :limit:`
	subQuery          = `(SELECT :select: FROM ":from:" WHERE :relation: :where: :groupBy: :orderBy: :limit:)`
	lateralQuery      = ":direction: JOIN LATERAL ( :query: ) :as: ON true"
	singleQuery       = "SELECT to_jsonb(q) FROM ( :query: ) q"
	sliceQuery        = "SELECT jsonb_agg(q) FROM ( :query: ) q"
	branchSingleQuery = `to_jsonb( :select: ) :as:`
	branchSliceQuery  = "jsonb_agg( :select: ) :as:"
	branchAnonQuery   = ":select:"
)

type (
	query struct {
		q string
	}
)

func newSliceQuery() *query {
	return &query{
		q: sliceQuery,
	}
}

func newSingleQuery() *query {
	return &query{
		q: singleQuery,
	}
}

func newRootQuery() *query {
	return &query{
		q: rootQuery,
	}
}

func newBaseQuery() *query {
	return &query{
		q: baseQuery,
	}
}

func newLateralQuery() *query {
	return &query{
		q: lateralQuery,
	}
}

func newBranchSingle() *query {
	return &query{
		q: branchSingleQuery,
	}
}

func newBranchSlice() *query {
	return &query{
		q: branchSliceQuery,
	}
}

func newBranchAnon() *query {
	return &query{
		q: branchAnonQuery,
	}
}

func (q *query) setSelect(value string) *query {
	q.q = strings.Replace(q.q, ":select:", value, 1)

	return q
}

func (q *query) setJoin(value string) *query {
	q.q = strings.Replace(q.q, ":join:", value, 1)

	return q
}
func (q *query) setQuery(value string) *query {
	q.q = strings.Replace(q.q, ":query:", value, 1)

	return q
}

func (q *query) setDirection(value string) *query {
	q.q = strings.Replace(q.q, ":direction:", value, 1)

	return q
}

func (q *query) setWhere(value string) *query {
	var where string

	if value != "" {
		where = fmt.Sprintf("WHERE %s ", value)
	}

	q.q = strings.Replace(q.q, ":where:", where, 1)

	return q
}

func (q *query) setWhereNulls(value string) *query {
	var where string

	if value != "" {
		where = fmt.Sprintf("WHERE %s ", value)
	}

	q.q = strings.Replace(q.q, ":whereNulls:", where, 1)

	return q
}

func (q *query) setFrom(value string) *query {
	q.q = strings.Replace(q.q, ":from:", value, 1)

	return q
}

func (q *query) setGroupBy(gb string) *query {
	str := ""
	if gb != "" {
		str = fmt.Sprintf("GROUP BY %s", gb)
	}

	q.q = strings.Replace(q.q, ":groupBy:", str, 1)

	return q
}

func (q *query) setOrderBy(ob string) *query {
	str := ""
	if ob != "" {
		str = fmt.Sprintf("ORDER BY %s", ob)
	}

	q.q = strings.Replace(q.q, ":orderBy:", str, 1)

	return q
}

func (q *query) setLimit(filters *Filters) *query {
	if filters == nil {
		return q
	}

	if filters.DisablePaging {
		return q
	}

	offset := 0
	if filters.Page > 1 {
		offset = (filters.Page - 1) * filters.PerPage
	}

	q.q = strings.Replace(q.q, ":limit:", fmt.Sprintf("LIMIT %d OFFSET %d ", filters.PerPage, offset), 1)

	return q
}

func (q *query) setAs(as string) *query {
	as = strings.TrimSpace(as)
	if as != "" {
		as = fmt.Sprintf(` AS "%s"`, as)
	}

	q.q = strings.Replace(q.q, ":as:", as, 1)

	return q
}

func (q *query) SetTotalRows(value string) *query {
	q.q = strings.Replace(q.q, ":totalRows:", value, 1)

	return q
}

var (
	scrubReplacers = regexp.MustCompile("(:[a-zA-z0-9]+:)")
	scrubSpaces    = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
)

func (q *query) Scrub() string {
	q.q = scrubReplacers.ReplaceAllString(q.q, "")
	q.q = scrubSpaces.ReplaceAllString(q.q, " ")

	return strings.TrimSpace(q.q)
}
