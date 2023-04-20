package liqu

import (
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

type Condition struct {
	Field      string
	Operator   Operator
	Value      interface{}
	NestedOp   Operator
	Conditions []*Condition
}

type WhereClause struct {
	Conditions []*Condition
}

func (wc *WhereClause) AddCondition(field string, op Operator, value interface{}) {
	wc.Conditions = append(wc.Conditions, &Condition{
		Field:    field,
		Operator: op,
		Value:    value,
	})
}

func (wc *WhereClause) AddNestedCondition(nestedOp Operator, conditions ...*Condition) {
	wc.Conditions = append(wc.Conditions, &Condition{
		NestedOp:   nestedOp,
		Conditions: conditions,
	})
}

func buildConditions(conditions []*Condition, defaultOp Operator) string {
	var sb strings.Builder

	for i, c := range conditions {
		if i > 0 && c.NestedOp == "" {
			sb.WriteString(" " + string(defaultOp) + " ")
		} else if i > 0 {
			sb.WriteString(" " + string(c.NestedOp) + " ")
		}

		if c.Conditions != nil {
			sb.WriteString("(")
			sb.WriteString(buildConditions(c.Conditions, c.NestedOp))
			sb.WriteString(")")
		} else {
			sb.WriteString(c.Field + " " + string(c.Operator) + " ")

			if c.Operator == IsNull || c.Operator == IsNotNull {
				continue
			}

			switch v := c.Value.(type) {
			case string:
				sb.WriteString("'" + v + "'")
			case int, float64:
				sb.WriteString(fmt.Sprintf("%v", v))
			case []interface{}:
				if c.Operator == In || c.Operator == Between || c.Operator == NotIn {
					values := make([]string, len(v))
					for i, value := range v {
						switch value := value.(type) {
						case string:
							values[i] = fmt.Sprintf("'%v'", value)
						default:
							values[i] = fmt.Sprintf("%v", value)
						}
					}
					sb.WriteString("(" + strings.Join(values, ", ") + ")")
				}
			}
		}
	}

	return sb.String()
}

func (wc *WhereClause) Build() string {
	return "WHERE " + buildConditions(wc.Conditions, And)
}
