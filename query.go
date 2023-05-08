package liqu

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	rootQuery         = "SELECT :totalRows: :select: FROM ( :from: ) :as: :join: :where: :groupBy: :orderBy: :limit:"
	baseQuery         = "SELECT :select: FROM :from: :as: :join: :where: :groupBy: :orderBy: :limit:"
	lateralQuery      = ":direction: JOIN LATERAL ( :query: ) :as: ON true"
	singleQuery       = "SELECT to_jsonb(q) FROM ( :query: ) q"
	sliceQuery        = "SELECT jsonb_agg(q)) FROM ( :query: ) q"
	branchSingleQuery = "to_jsonb( :select: ) :as:"
	branchSliceQuery  = "coalesce(jsonb_agg( :select: ), '[]') :as:"
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

func (q *query) setFrom(value string) *query {
	q.q = strings.Replace(q.q, ":from:", value, 1)

	return q
}

func (q *query) setGroupBy(groupBy []string) *query {
	groupByString := ""
	if len(groupBy) > 0 {
		groupByString = " GROUP BY " + strings.Join(groupBy, ",")
	}

	q.q = strings.Replace(q.q, ":groupBy:", groupByString, 1)

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

func (q *query) setLimit(paging *Paging) *query {
	if paging == nil {
		return q
	}

	if paging.Disabled {
		return q
	}

	offset := 0
	if paging.Page > 1 {
		offset = (paging.Page - 1) * paging.PerPage
	}

	q.q = strings.Replace(q.q, ":limit:", fmt.Sprintf("LIMIT %d OFFSET %d ", paging.PerPage, offset), 1)

	return q
}

func (q *query) setAs(this string, that string) *query {
	as := strings.TrimSpace(this)

	if as == "" {
		as = strings.TrimSpace(that)
	}

	if as != "" {
		as = fmt.Sprintf(" AS %s", as)
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
