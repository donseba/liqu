package liqu

import (
	"errors"
	"fmt"
	"strings"
)

type Operator string

const (
	And                Operator = "AND"
	Or                 Operator = "OR"
	Equal              Operator = "="
	NotEqual           Operator = "<>"
	NotEqualAlt        Operator = "!="
	LessThan           Operator = "<"
	LessThanOrEqual    Operator = "<="
	GreaterThan        Operator = ">"
	GreaterThanOrEqual Operator = ">="
	Like               Operator = "LIKE"
	ILike              Operator = "ILIKE"
	NotLike            Operator = "NOT LIKE"
	NotILike           Operator = "NOT ILIKE"
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
	column     string
	conditions []string
}

// NewConditionBuilder initializes and returns a new ConditionBuilder
func NewConditionBuilder() *ConditionBuilder {
	return &ConditionBuilder{
		conditions: []string{},
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

	switch v := value.(type) {
	case nil:
		condition = fmt.Sprintf("%s %s", cb.column, op)
	case string:
		condition = fmt.Sprintf("%s %s '%s'", cb.column, op, v)
	default:
		condition = fmt.Sprintf("%s %s %v", cb.column, op, v)
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
		cb.conditions = append(cb.conditions, "AND")
	}
	return cb.Column(column).Condition(IsNotNull, nil)
}

// OrIsNull adds an OR condition with the IS NULL operator
func (cb *ConditionBuilder) OrIsNull(column string) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, "OR")
	}
	return cb.Column(column).Condition(IsNull, nil)
}

// OrIsNotNull adds an OR condition with the IS NOT NULL operator
func (cb *ConditionBuilder) OrIsNotNull(column string) *ConditionBuilder {
	if len(cb.conditions) > 0 {
		cb.conditions = append(cb.conditions, "OR")
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
	nestedCb := NewConditionBuilder()
	fn(nestedCb)
	nestedConditions := nestedCb.Build()

	if len(nestedConditions) > 0 {
		cb.conditions = append(cb.conditions, fmt.Sprintf("(%s)", nestedConditions))
	}

	return cb
}

// In, NotIn, Any, and NotAny methods with a variable number of arguments
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

// OrIn, OrNotIn, OrAny, and OrNotAny methods
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

// AndIn, AndNotIn, AndAny, and AndNotAny methods
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
	valueStrings := make([]string, len(values))
	for i, value := range values {
		switch v := value.(type) {
		case string:
			valueStrings[i] = fmt.Sprintf("'%s'", v)
		default:
			valueStrings[i] = fmt.Sprintf("%v", v)
		}
	}
	condition := fmt.Sprintf("%s %s (%s)", column, op, strings.Join(valueStrings, ", "))
	cb.conditions = append(cb.conditions, condition)
	return cb
}

// Build returns the final SQL WHERE clause
func (cb *ConditionBuilder) Build() string {
	return strings.Join(cb.conditions, " ")
}

func ParseURLQueryToConditionBuilder(query string) (*ConditionBuilder, error) {
	parts := strings.Split(query, ",")

	cb := NewConditionBuilder()

	for _, part := range parts {
		element := strings.Split(part, "|")

		column := element[0]
		operator := Operator(element[1])

		switch operator {
		case Equal, NotEqual, NotEqualAlt, LessThan, LessThanOrEqual, GreaterThan, GreaterThanOrEqual, Like, ILike, NotLike, NotILike, StartsWith:
			if len(element) < 3 {
				return nil, errors.New("invalid query format")
			}
			value := element[2]
			cb.And(column, operator, value)
		case In, NotIn, Any, NotAny:
			if len(element) < 3 {
				return nil, errors.New("invalid query format")
			}
			values := strings.Split(element[2], "--")
			interfaceValues := make([]interface{}, len(values))
			for i, v := range values {
				interfaceValues[i] = v
			}
			if operator == In {
				cb.AndIn(column, interfaceValues...)
			} else if operator == NotIn {
				cb.AndNotIn(column, interfaceValues...)
			} else if operator == Any {
				cb.AndAny(column, interfaceValues...)
			} else {
				cb.AndNotAny(column, interfaceValues...)
			}
		case IsNull, IsNotNull:
			if len(element) != 2 {
				return nil, errors.New("invalid query format")
			}
			cb.And(column, operator, nil)
		default:
			return nil, errors.New("unsupported operator")
		}
	}

	return cb, nil
}
