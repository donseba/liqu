package liqu

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
)

var (
	DefaultPage    = 1
	DefaultPerPage = 25
	regexTotalRows = regexp.MustCompile(`"totalrows":[0-9]+,`)
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
		paging      *Paging
		filters     *Filters

		sqlQuery  string
		sqlParams []interface{}
	}

	Filters struct {
		Page          int
		PerPage       int
		DisablePaging bool
		Where         string
		OrderBy       string
	}

	Paging struct {
		Page         int
		PerPage      int
		TotalResults int
		TotalPages   int
		Disabled     bool
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
	var (
		page     = DefaultPage
		perPage  = DefaultPerPage
		disabled = false
	)

	if filters != nil {
		page = filters.Page
		perPage = filters.PerPage
		disabled = filters.DisablePaging
	}

	return &Liqu{
		ctx:      ctx,
		registry: make(map[string]registry, 0),
		filters:  filters,
		paging: &Paging{
			Page:     page,
			PerPage:  perPage,
			Disabled: disabled,
		},
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

func (l *Liqu) Paging() *Paging {
	return l.paging
}

func (l *Liqu) PostProcess(pp string) string {
	var count int

	rexMatch := regexTotalRows.FindStringSubmatch(pp)
	if len(rexMatch) > 0 {
		pp = regexTotalRows.ReplaceAllString(pp, "")
		count, _ = strconv.Atoi(regexp.MustCompile("[0-9]+").FindString(rexMatch[0]))
	}

	l.paging.TotalResults = count
	l.paging.TotalPages = int(math.Ceil(float64(l.paging.TotalResults) / float64(l.paging.PerPage)))

	return pp
}

func ParseUrlValuesToFilters(values map[string]string) (*Filters, error) {
	filters := &Filters{
		Page:    DefaultPage,
		PerPage: DefaultPerPage,
	}

	if whereQuery, ok := values["where"]; ok {
		filters.Where = whereQuery
	}

	if orderQuery, ok := values["order"]; ok {
		filters.OrderBy = orderQuery
	}

	if pageQuery, ok := values["page"]; ok {
		pageInt, err := strconv.Atoi(pageQuery)
		if err != nil {
			return filters, err
		}

		filters.Page = pageInt
	}

	if perPageQuery, ok := values["per_page"]; ok {
		perPageInt, err := strconv.Atoi(perPageQuery)
		if err != nil {
			return filters, err
		}

		filters.PerPage = perPageInt
	}

	return filters, nil
}

func (l *Liqu) parseFilters() error {
	var (
		where string
		order string
	)
	if l.filters != nil {
		where = l.filters.Where
		order = l.filters.OrderBy
	}

	err := l.parseNestedConditions(where, l.tree.where, And)
	if err != nil {
		return err
	}

	err = l.parseOrderBy(order, l.tree.order)
	if err != nil {
		return err
	}

	return nil
}
