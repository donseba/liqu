package liqu

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
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

		sqlQuery  string
		sqlParams []interface{}
	}

	Filters struct {
	}

	registry struct {
		fieldTypes    map[string]reflect.Type
		fieldDatabase map[string]string

		//// SubQueryFields
		//SubQueryField map[string]string
		tableName string
		branch    *branch
	}
)

func New(ctx context.Context, filters *Filters) *Liqu {
	return &Liqu{
		ctx:      ctx,
		registry: make(map[string]registry, 0),
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

func (l *Liqu) PostProcess(pp string) string {
	var count int

	rexMatch := regexTotalRows.FindStringSubmatch(pp)
	if len(rexMatch) > 0 {
		pp = regexTotalRows.ReplaceAllString(pp, "")
		count, _ = strconv.Atoi(regexp.MustCompile("[0-9]+").FindString(rexMatch[0]))
	}

	_ = count
	//l.paging.TotalResults = count
	//l.paging.TotalPages = int(math.Ceil(float64(l.paging.TotalResults) / float64(l.paging.PerPage)))

	return pp
}

func printStructuredSQL(sqlQuery string) {
	// Define the keywords for indentation
	keywords := []string{"SELECT", "FROM", "LEFT JOIN", "ON", "AS", ")", "(", ","}

	// Replace newline characters with spaces and split the SQL query into words
	words := strings.Split(strings.ReplaceAll(sqlQuery, "\n", " "), " ")

	indent := 0
	for _, word := range words {
		// Check if the word is a keyword
		isKeyword := false
		for _, keyword := range keywords {
			if strings.HasPrefix(strings.ToUpper(word), keyword) {
				isKeyword = true
				break
			}
		}

		// If the word is a keyword, print it with the current indentation
		if isKeyword {
			if strings.HasPrefix(strings.ToUpper(word), "(") || strings.HasPrefix(strings.ToUpper(word), ",") {
				fmt.Printf("\n%s%s", strings.Repeat("  ", indent), word)
			} else {
				fmt.Printf("\n%s%s", strings.Repeat("  ", indent-1), word)
			}

			// Adjust the indent level based on the keyword
			if strings.HasPrefix(strings.ToUpper(word), "SELECT") ||
				strings.HasPrefix(strings.ToUpper(word), "FROM") ||
				strings.HasPrefix(strings.ToUpper(word), "LEFT JOIN") ||
				strings.HasPrefix(strings.ToUpper(word), "(") {
				indent++
			} else if strings.HasPrefix(strings.ToUpper(word), ")") ||
				strings.HasPrefix(strings.ToUpper(word), "ON") {
				indent--
			}
		} else {
			// Otherwise, print the word with a space before it
			fmt.Printf(" %s", word)
		}
	}
	fmt.Println()
}
