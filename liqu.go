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
	"strings"
)

var (
	DefaultPage    = 1
	DefaultPerPage = 25
	regexTotalRows = regexp.MustCompile(`,?\s?"totalrows":\s?[0-9]+,?`)
)

type (
	Source interface {
		Table() string
		PrimaryKeys() []string
	}

	Liqu struct {
		source             interface{}
		sourceType         reflect.Type
		sourceSlice        bool
		tree               *branch
		registry           map[string]registry
		linkedCte          map[string][]linkedCte
		filters            *Filters
		defaults           *Defaults
		subQueries         []*SubQuery
		cte                map[string]*Cte
		cteBranchedQueries []*CteBranchedQuery

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
		registry:   make(map[string]registry, 0),
		linkedCte:  make(map[string][]linkedCte, 0),
		cte:        make(map[string]*Cte),
		filters:    filters,
		defaults:   NewDefaults(),
		subQueries: make([]*SubQuery, 0),
	}
}

func (l *Liqu) WithDefaults(defaults *Defaults) *Liqu {
	l.defaults = defaults

	return l
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

	err = l.parseSubQueries()
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
		var replacer string
		if strings.HasPrefix(rexMatch[0], ",") && strings.HasSuffix(rexMatch[0], ",") {
			replacer = ", "
		}
		pp = regexTotalRows.ReplaceAllString(pp, replacer)
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
		PushUrl: true,
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

	if orderQuery, ok := values["order_by"]; ok {
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

	if pushUrlQuery, ok := values["push_url"]; ok {
		if len(pushUrlQuery) > 0 {
			pushURL, _ := strconv.ParseBool(pushUrlQuery[0])
			filters.PushUrl = pushURL
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

	err := l.processDefaults()
	if err != nil {
		return err
	}

	err = l.parseNestedConditions(where, l.tree.where, And)
	if err != nil {
		return err
	}

	err = l.parseOrderBy(order)
	if err != nil {
		return err
	}

	err = l.parseSelect(sel, true)
	if err != nil {
		return err
	}

	return nil
}
