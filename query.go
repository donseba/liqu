package liqu

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	rootQuery    = "SELECT :totalRows: :select: FROM ( :from: ) :as: :join: :where: :groupBy: :orderBy: :limit:"
	baseQuery    = "SELECT :select: FROM :from: :as: :join: :where: :groupBy: :limit:"
	lateralQuery = ":direction: JOIN LATERAL ( :baseQuery: ) :as: ON true"
	sliceQuery   = "SELECT array_to_json(array_agg(q)) FROM ( :query: ) q"
	singleQuery  = "SELECT row_to_json(q) FROM ( :query: ) q"
)

type (
	query struct {
		q string
	}
)

func NewRootQuery() *query {
	return &query{
		q: rootQuery,
	}
}
func NewBaseQuery() *query {
	return &query{
		q: baseQuery,
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

func (q *query) setOrderBy(ob []string) *query {
	str := ""
	if len(ob) > 0 {
		str = "ORDER BY " + strings.Join(ob, ", ")
	}

	q.q = strings.Replace(q.q, ":orderBy:", str, 1)

	return q
}

//func (q *query) setLimit(paging *Paging) *query {
//	if paging == nil {
//		return strings.Replace(q, ":limit:", "", 1)
//	}
//
//	if paging.Disabled {
//		return strings.Replace(q, ":limit:", "", 1)
//	}
//
//	offset := 0
//	if paging.Page > 1 {
//		offset = (paging.Page - 1) * paging.PerPage
//	}
//
//	return strings.Replace(q, ":limit:", fmt.Sprintf("LIMIT %d OFFSET %d ", paging.PerPage, offset), 1)
//}

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
