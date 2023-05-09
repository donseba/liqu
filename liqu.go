package liqu

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
)

var (
	DefaultPage    = 1
	DefaultPerPage = 25
	regexTotalRows = regexp.MustCompile(`,?\s?"totalrows":\s?[0-9]+,?\s?`)
)

type (
	Source interface {
		Table() string
		PrimaryKeys() []string
	}

	Liqu struct {
		ctx         context.Context
		source      interface{}
		sourceType  reflect.Type
		sourceSlice bool
		tree        *branch
		registry    map[string]registry
		filters     *Filters

		sqlQuery  string
		sqlParams []interface{}
	}

	registry struct {
		fieldTypes    map[string]reflect.Type
		fieldDatabase map[string]string
		fieldSearch   map[string]interface{}
		tableName     string
		branch        *branch
	}
)

func New(ctx context.Context, filters *Filters) *Liqu {
	if filters == nil {
		filters = &Filters{
			Page:          DefaultPage,
			PerPage:       DefaultPerPage,
			DisablePaging: false,
		}
	}

	if filters.Page <= 0 {
		filters.Page = DefaultPage
	}

	if filters.PerPage <= 0 {
		filters.PerPage = DefaultPerPage
	}

	return &Liqu{
		ctx:      ctx,
		registry: make(map[string]registry, 0),
		filters:  filters,
	}
}

func (l *Liqu) FromSource(source interface{}) error {
	var (
		sourceType  = reflect.ValueOf(source).Type()
		sourceSlice = false
	)

	// get the root element and check for slice
	sourceType, sourceSlice = getRootElement(sourceType)

	// we need a struct at this point
	if sourceType.Kind() != reflect.Struct {
		return fmt.Errorf("[liqu] source needs to be a struct, slice of structs or slice of struct pointer, got %+v instead", sourceType.Kind())
	}

	if sourceType.NumField() == 0 {
		return errors.New("[liqu] source needs to have at least one field")
	}

	// set everything we know about the source
	l.source = source
	l.sourceType = sourceType
	l.sourceSlice = sourceSlice

	err := l.scan(l.sourceType, nil)
	if err != nil {
		return err
	}

	err = l.parseFilters()
	if err != nil {
		return err
	}

	err = l.traverse()
	if err != nil {
		return err
	}

	return nil
}

func getRootElement(rt reflect.Type) (reflect.Type, bool) {
	var slice bool
	// pointer check
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	// slice check
	if rt.Kind() == reflect.Slice {
		slice = true
		rt = rt.Elem()
	}

	// pointer check in array
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	return rt, slice
}

func Debug(v ...interface{}) {
	fmt.Println("-------------")

	_, fn, line, _ := runtime.Caller(1)
	fmt.Printf("[debug] %s:%d \n", fn, line)

	for i := 0; i < len(v); i++ {
		fmt.Printf("%+v\n", v[i])
	}
	fmt.Println("-------------")
}

func (l *Liqu) SQL() (string, []interface{}) {
	return l.sqlQuery, l.sqlParams
}

func (l *Liqu) Filters() *Filters {
	return l.filters
}

func (l *Liqu) PostProcess(pp string) string {
	var count int

	rexMatch := regexTotalRows.FindStringSubmatch(pp)
	if len(rexMatch) > 0 {
		pp = regexTotalRows.ReplaceAllString(pp, "")
		count, _ = strconv.Atoi(regexp.MustCompile("[0-9]+").FindString(rexMatch[0]))
	}

	l.filters.totalResults = count
	l.filters.totalPages = int(math.Ceil(float64(l.filters.totalResults) / float64(l.filters.PerPage)))

	return pp
}

func ParseUrlValuesToFilters(values url.Values) (*Filters, error) {
	filters := &Filters{
		Page:    DefaultPage,
		PerPage: DefaultPerPage,
	}

	if selectQuery, ok := values["select"]; ok {
		if len(selectQuery) > 0 {
			filters.Select = selectQuery[0]
		}
	}
	if whereQuery, ok := values["where"]; ok {
		if len(whereQuery) > 0 {
			filters.Where = whereQuery[0]
		}
	}

	if orderQuery, ok := values["order"]; ok {
		if len(orderQuery) > 0 {
			filters.OrderBy = orderQuery[0]
		}
	}

	if pageQuery, ok := values["page"]; ok {
		if len(pageQuery) > 0 {
			pageInt, err := strconv.Atoi(pageQuery[0])
			if err != nil {
				return filters, err
			}

			filters.Page = pageInt
		}
	}

	if perPageQuery, ok := values["per_page"]; ok {
		if len(perPageQuery) > 0 {
			perPageInt, err := strconv.Atoi(perPageQuery[0])
			if err != nil {
				return filters, err
			}

			filters.PerPage = perPageInt
		}
	}

	return filters, nil
}

func (l *Liqu) parseFilters() error {
	var (
		where string
		order string
		sel   string
	)
	if l.filters != nil {
		where = l.filters.Where
		order = l.filters.OrderBy
		sel = l.filters.Select
	}

	err := l.parseNestedConditions(where, l.tree.where, And)
	if err != nil {
		return err
	}

	err = l.parseOrderBy(order)
	if err != nil {
		return err
	}

	err = l.parseSelect(sel)
	if err != nil {
		return err
	}

	return nil
}
