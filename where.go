package liqu

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Operator string

const (
	And                Operator = "AND"
	Or                 Operator = "OR"
	Equal              Operator = "="
	NotEqual           Operator = "<>"
	LessThan           Operator = "<"
	LessThanOrEqual    Operator = "<="
	GreaterThan        Operator = ">"
	GreaterThanOrEqual Operator = ">="
	Like               Operator = "~~"
	ILike              Operator = "~~*"
	NotLike            Operator = "!~~"
	NotILike           Operator = "!~~*"
	In                 Operator = "IN"
	Between            Operator = "BETWEEN"
	NotIn              Operator = "NOT IN"
	Any                Operator = "ANY"
	NotAny             Operator = "NOT ANY"
	StartsWith         Operator = "^"
	IsNull             Operator = "IS NULL"
	IsNotNull          Operator = "IS NOT NULL"
)

func (o Operator) String() string {
	return string(o)
}

// ConditionBuilder is a struct for fluently building SQL WHERE clauses
type ConditionBuilder struct {
	column           string
	conditions       []string
	args             []any
	counter          int
	liqu             *Liqu
	protectedColumns map[string]bool
}

// NewConditionBuilder initializes and returns a new ConditionBuilder
func NewConditionBuilder() *ConditionBuilder {
	return &ConditionBuilder{
		conditions:       []string{},
		args:             []any{},
		counter:          0,
		protectedColumns: make(map[string]bool),
	}
}

// Column sets the column for the condition
func (cb *ConditionBuilder) Column(column string) *ConditionBuilder {
	cb.column = column
	return cb
}

// Condition adds a condition with the provided operator and value
func (cb *ConditionBuilder) Condition(op Operator, value interface{}) *ConditionBuilder {
	var condition string

	if queried, ok := cb.protectedColumns[cb.column]; ok && queried {
		return cb
	}

	if value != nil && reflect.TypeOf(value).Kind() == reflect.Slice {
		slice, ok := reflect.ValueOf(value).Interface().([]string)
		if !ok {
			panic("value not a []string")
		}

		var values []any
		for _, v := range slice {
			values = append(values, any(v))
		}

		return cb.multiValueCondition(cb.column, op, values)
	} else if op == In || op == NotIn || op == Any || op == NotAny {
		return cb.multiValueCondition(cb.column, op, []interface{}{value})
	}

	if value == nil {
		condition = fmt.Sprintf("%s %s", cb.column, op)
	} else {
		cb.args = append(cb.args, value)
		cb.counter++
		if cb.liqu != nil {
			cb.liqu.sqlParams = append(cb.liqu.sqlParams, value)
			condition = fmt.Sprintf("%s %s $%d", cb.column, op, len(cb.liqu.sqlParams))
		} else {
			condition = fmt.Sprintf("%s %s $%d", cb.column, op, cb.counter)
		}
	}

	cb.conditions = append(cb.conditions, condition)
	return cb
}

// And adds an AND condition with the provided column, operator, and value
func (cb *ConditionBuilder) And(column string, op Operator, value interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, And.String())
	}
	return cb.Column(column).Condition(op, value)
}

func (cb *ConditionBuilder) AndRaw(raw string) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, And.String())
	}

	cb.conditions = append(cb.conditions, raw)

	return cb
}

// Or adds an OR condition with the provided column, operator, and value
func (cb *ConditionBuilder) Or(column string, op Operator, value interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, Or.String())
	}
	return cb.Column(column).Condition(op, value)
}

// AndIsNull adds an AND condition with the IS NULL operator
func (cb *ConditionBuilder) AndIsNull(column string) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, And.String())
	}
	return cb.Column(column).Condition(IsNull, nil)
}

// AndIsNotNull adds an AND condition with the IS NOT NULL operator
func (cb *ConditionBuilder) AndIsNotNull(column string) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, And.String())
	}
	return cb.Column(column).Condition(IsNotNull, nil)
}

// OrIsNull adds an OR condition with the IS NULL operator
func (cb *ConditionBuilder) OrIsNull(column string) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, Or.String())
	}
	return cb.Column(column).Condition(IsNull, nil)
}

// OrIsNotNull adds an OR condition with the IS NOT NULL operator
func (cb *ConditionBuilder) OrIsNotNull(column string) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, Or.String())
	}
	return cb.Column(column).Condition(IsNotNull, nil)
}

// AndNested adds a nested set of AND conditions using the provided function
func (cb *ConditionBuilder) AndNested(fn func(*ConditionBuilder)) *ConditionBuilder {
	cb.conditions = append(cb.conditions, And.String())

	return cb.Nested(fn)
}

// OrNested adds a nested set of OR conditions using the provided function
func (cb *ConditionBuilder) OrNested(fn func(*ConditionBuilder)) *ConditionBuilder {
	cb.conditions = append(cb.conditions, Or.String())
	return cb.Nested(fn)
}

// Nested adds a nested set of conditions using the provided function
func (cb *ConditionBuilder) Nested(fn func(*ConditionBuilder)) *ConditionBuilder {
	nestedCb := NewConditionBuilder().setCounter(cb.counter).setLiqu(cb.liqu)
	fn(nestedCb)

	nestedConditions := nestedCb.Build()
	nestedArgs := nestedCb.Args()

	if len(nestedConditions) > 0 {
		cb.conditions = append(cb.conditions, fmt.Sprintf("(%s)", nestedConditions))
		cb.args = append(cb.args, nestedArgs...)
	}

	return cb.setCounter(nestedCb.counter)
}

func (cb *ConditionBuilder) In(column string, values ...interface{}) *ConditionBuilder {
	return cb.multiValueCondition(column, In, values)
}

func (cb *ConditionBuilder) NotIn(column string, values ...interface{}) *ConditionBuilder {
	return cb.multiValueCondition(column, NotIn, values)
}

func (cb *ConditionBuilder) Any(column string, values ...interface{}) *ConditionBuilder {
	return cb.multiValueCondition(column, Any, values)
}

func (cb *ConditionBuilder) NotAny(column string, values ...interface{}) *ConditionBuilder {
	return cb.multiValueCondition(column, NotAny, values)
}

func (cb *ConditionBuilder) OrIn(column string, values ...interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, Or.String())
	}
	return cb.In(column, values...)
}

func (cb *ConditionBuilder) OrNotIn(column string, values ...interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, Or.String())
	}
	return cb.NotIn(column, values...)
}

func (cb *ConditionBuilder) OrAny(column string, values ...interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, Or.String())
	}
	return cb.Any(column, values...)
}

func (cb *ConditionBuilder) OrNotAny(column string, values ...interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, Or.String())
	}
	return cb.NotAny(column, values...)
}

func (cb *ConditionBuilder) AndIn(column string, values ...interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, And.String())
	}
	return cb.In(column, values...)
}

func (cb *ConditionBuilder) AndNotIn(column string, values ...interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, And.String())
	}
	return cb.NotIn(column, values...)
}

func (cb *ConditionBuilder) AndAny(column string, values ...interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, And.String())
	}
	return cb.Any(column, values...)
}

func (cb *ConditionBuilder) AndNotAny(column string, values ...interface{}) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, And.String())
	}
	return cb.NotAny(column, values...)
}

func (cb *ConditionBuilder) multiValueCondition(column string, op Operator, values []interface{}) *ConditionBuilder {
	if op == Any || op == NotAny {
		cb.counter++
		placeholder := fmt.Sprintf("$%d", cb.counter)

		condition := fmt.Sprintf("%s && %s", column, placeholder)
		if op == NotAny {
			condition = fmt.Sprintf("NOT(%s && %s)", column, placeholder)
		}
		cb.conditions = append(cb.conditions, condition)
		cb.args = append(cb.args, values)
		if cb.liqu != nil {
			cb.liqu.sqlParams = append(cb.liqu.sqlParams, array(values))
		}
		return cb
	}

	placeholders := make([]string, len(values))
	for i, value := range values {
		if cb.liqu != nil {
			cb.counter++
			cb.liqu.sqlParams = append(cb.liqu.sqlParams, value)
			placeholders[i] = fmt.Sprintf("$%d", len(cb.liqu.sqlParams))
		} else {
			cb.counter++
			placeholders[i] = fmt.Sprintf("$%d", cb.counter)
		}
	}

	condition := fmt.Sprintf("%s %s (%s)", column, op, strings.Join(placeholders, ", "))
	cb.conditions = append(cb.conditions, condition)
	cb.args = append(cb.args, values...)

	return cb
}

// Build returns the final SQL WHERE clause
func (cb *ConditionBuilder) Build() string {
	return strings.Join(cb.conditions, " ")
}

func (cb *ConditionBuilder) Args() []interface{} {
	return cb.args
}

func (cb *ConditionBuilder) setCounter(c int) *ConditionBuilder {
	cb.counter = c

	return cb
}

func (cb *ConditionBuilder) setLiqu(liqu *Liqu) *ConditionBuilder {
	cb.liqu = liqu

	return cb
}

func ParseURLQueryToConditionBuilder(query string) (*ConditionBuilder, error) {
	cb := NewConditionBuilder()
	return parseNestedConditions(query, cb, And)
}

func parseNestedConditions(query string, cb *ConditionBuilder, outerOperator Operator) (*ConditionBuilder, error) {
	parts := strings.Split(query, ",")

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		element := strings.Split(part, "|")

		if strings.HasPrefix(part, "(") {
			nestedOperator := Operator(element[0][1:])
			if nestedOperator != And && nestedOperator != Or {
				return nil, errors.New("invalid nested operator")
			}

			i++
			nestedQuery := ""
			nestedCount := 1

			for ; i < len(parts) && nestedCount > 0; i++ {
				if strings.HasPrefix(parts[i], "(") {
					nestedCount++
				} else if strings.HasSuffix(parts[i], ")") {
					nestedCount--
				}
				nestedQuery += parts[i] + ","
			}

			// Remove trailing comma
			nestedQuery = strings.TrimSuffix(nestedQuery, ",")

			if nestedCount != 0 {
				return nil, errors.New("unbalanced parentheses")
			} else {
				i--
			}

			// Remove the last closing parenthesis
			nestedQuery = strings.TrimSuffix(nestedQuery, ")")

			var err error
			if outerOperator == And {
				cb.AndNested(func(nestedCB *ConditionBuilder) {
					_, err = parseNestedConditions(nestedQuery, nestedCB, nestedOperator)
				})
			} else {
				cb.OrNested(func(nestedCB *ConditionBuilder) {
					_, err = parseNestedConditions(nestedQuery, nestedCB, nestedOperator)
				})
			}
			if err != nil {
				return nil, fmt.Errorf("[liqu] error in nested query: %s", err.Error())
			}
		} else {
			if len(element) < 2 {
				return nil, fmt.Errorf("invalid query format: %s", part)
			}

			column := element[0]
			operator := Operator(element[1])

			if len(element) == 3 {
				var value interface{}
				value = element[2]
				if strings.Contains(element[2], "--") {
					value = strings.Split(element[2], "--")
				} else {
					value = element[2]
				}

				if outerOperator == And {
					cb.And(column, operator, value)
				} else {
					cb.Or(column, operator, value)
				}
			} else {
				if outerOperator == And {
					cb.And(column, operator, nil)
				} else {
					cb.Or(column, operator, nil)
				}
			}
		}
	}

	return cb, nil
}

func (l *Liqu) parseNestedConditions(query string, cb *ConditionBuilder, outerOperator Operator) error {
	if strings.TrimSpace(query) == "" {
		return nil
	}

	parts := strings.Split(query, ",")

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		element := strings.Split(part, "|")

		if strings.HasPrefix(part, "(") {
			nestedOperator := Operator(element[0][1:])
			if nestedOperator != And && nestedOperator != Or {
				return errors.New("invalid nested operator")
			}

			i++
			nestedQuery := ""
			nestedCount := 1

			for ; i < len(parts) && nestedCount > 0; i++ {
				if strings.HasPrefix(parts[i], "(") {
					nestedCount++
				} else if strings.HasSuffix(parts[i], ")") {
					nestedCount--
				}
				nestedQuery += parts[i] + ","
			}

			// Remove trailing comma
			nestedQuery = strings.TrimSuffix(nestedQuery, ",")

			if nestedCount != 0 {
				return errors.New("unbalanced parentheses")
			} else {
				i--
			}

			// Remove the last closing parenthesis
			nestedQuery = strings.TrimSuffix(nestedQuery, ")")

			var err error
			if outerOperator == And {
				cb.AndNested(func(nestedCB *ConditionBuilder) {
					err = l.parseNestedConditions(nestedQuery, nestedCB, nestedOperator)
				})
			} else {
				cb.OrNested(func(nestedCB *ConditionBuilder) {
					err = l.parseNestedConditions(nestedQuery, nestedCB, nestedOperator)
				})
			}
			if err != nil {
				return fmt.Errorf("[liqu] error in nested query: %s", err.Error())
			}
		} else {
			if len(element) < 2 {
				return fmt.Errorf("invalid query format: %s", part)
			}

			el3 := ""
			if len(element) == 3 {
				el3 = element[2]
			}

			err := l.processWhere(outerOperator, element[0], element[1], el3, false)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Liqu) processWhere(outerOperator Operator, col string, op string, val interface{}, protect bool) error {
	var (
		cteTable string
		model    string
		field    string
		column   string
	)

	if strings.Contains(col, ".") {
		el := strings.Split(col, ".")
		model = el[0]
		field = el[1]

		if strings.Contains(model, "--") {
			cp := strings.Split(model, "--")
			model = cp[0]
			cteTable = cp[1]
		}
	} else {
		model = l.tree.as
		field = col
	}

	var ok bool

	if column, ok = l.registry[model].fieldDatabase[field]; !ok {
		return fmt.Errorf("invalid search field %s", col)
	}

	if l.registry[model].branch.isCTE {
		if cteTable == "" {
			return fmt.Errorf("provide cteTable for cte search %s", col)
		}

		l.cte[model].isSearched = true
		column = fmt.Sprintf(`"%s"."%s"`, l.registry[model].tableName, column)
	} else {
		column = fmt.Sprintf(`"%s"."%s"`, l.registry[model].tableName, column)
	}

	operator := Operator(op)

	l.registry[model].branch.selectedFields[field] = true
	l.registry[model].branch.isSearched = true

	if val != nil {
		var value interface{}
		value = val
		sval := fmt.Sprintf("%s", val)
		if strings.Contains(sval, "--") {
			value = strings.Split(sval, "--")
		} else {
			value = val
		}

		if outerOperator == And {
			l.registry[model].branch.where.And(column, operator, value)
		} else {
			l.registry[model].branch.where.Or(column, operator, value)
		}
	} else {
		if outerOperator == And {
			l.registry[model].branch.where.And(column, operator, nil)
		} else {
			l.registry[model].branch.where.Or(column, operator, nil)
		}
	}

	if protect {
		l.registry[model].branch.where.ProtectColumn(column)
	}

	return nil
}

func (cb *ConditionBuilder) ProtectColumn(column string) *ConditionBuilder {
	if cb.protectedColumns == nil {
		cb.protectedColumns = make(map[string]bool)
	}
	cb.protectedColumns[column] = true
	return cb
}

type array []interface{}

func (a array) Value() (string, error) {
	if len(a) == 0 {
		return "{}", nil
	}

	result := make([]string, len(a))
	for i, v := range a {
		str, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("unsupported type: %T", v)
		}

		// Try int
		if _, err := strconv.ParseInt(str, 10, 64); err == nil {
			result[i] = str
			continue
		}

		// Try float
		if _, err := strconv.ParseFloat(str, 64); err == nil {
			result[i] = str
			continue
		}

		// Try bool
		if _, err := strconv.ParseBool(str); err == nil {
			result[i] = str
			continue
		}

		// If all else fails, treat as string
		result[i] = "'" + strings.Replace(str, "'", "''", -1) + "'"
	}
	return "{" + strings.Join(result, ",") + "}", nil
}
