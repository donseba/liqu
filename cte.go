package liqu

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type (
	Cte struct {
		as          string
		source      interface{}
		sourceType  reflect.Type
		sourceSlice bool
		isSearched  bool
		baseTable   string

		liqu          *Liqu
		model         string
		selects       []string
		fieldDatabase map[string]string
		joins         []string
		orderBy       *OrderBuilder
		conditions    *ConditionBuilder
		groupBy       *GroupByBuilder
	}

	linkTrigger string
)

const (
	LinkAlways linkTrigger = "Always"
	LinkSearch linkTrigger = "Search"
)

func (l *Liqu) WithCte(as string, cte *Cte) *Liqu {
	cte.as = as

	l.cte[as] = cte

	l.registry[as] = registry{
		fieldDatabase: cte.fieldDatabase,
		tableName:     cte.baseTable,
		branch: &branch{
			isCTE:          true,
			selectedFields: map[string]bool{},
			where:          cte.conditions,
		},
	}

	return l
}

func NewCTE(li *Liqu, source interface{}) (*Cte, error) {
	var (
		sourceType  = reflect.ValueOf(source).Type()
		sourceSlice = false
	)

	// get the root element and check for slice
	sourceType, sourceSlice = getRootElement(sourceType)

	// we need a struct at this point
	if sourceType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("[liqu][cte] source needs to be a struct, slice of structs or slice of struct pointer, got %+v instead", sourceType.Kind())
	}

	if sourceType.NumField() == 0 {
		return nil, errors.New("[liqu][cte] source needs to have at least one field")
	}

	cte := &Cte{
		liqu:          li,
		source:        source,
		sourceType:    sourceType,
		sourceSlice:   sourceSlice,
		selects:       make([]string, 0),
		joins:         make([]string, 0),
		fieldDatabase: make(map[string]string),
		orderBy:       NewOrderBuilder(),
		conditions:    NewConditionBuilder().setLiqu(li),
		groupBy:       NewGroupByBuilder(),
	}

	err := cte.scan(cte.sourceType, nil)
	if err != nil {
		return cte, err
	}

	return cte, nil
}

func (cte *Cte) Select(field string) *Cte {
	cte.selects = append(cte.selects, field)
	return cte
}

func (cte *Cte) SelectAs(field, alias string) *Cte {
	cte.selects = append(cte.selects, fmt.Sprintf("%s AS %s", field, alias))
	return cte
}

func (cte *Cte) SelectAggregate(funcName, field, alias string) *Cte {
	cte.selects = append(cte.selects, fmt.Sprintf("%s(%s) AS %s", funcName, field, alias))
	return cte
}

func (cte *Cte) Join(joinTable, condition, joinType string) *Cte {
	join := fmt.Sprintf("%s JOIN %s ON %s", strings.ToUpper(joinType), joinTable, condition)
	cte.joins = append(cte.joins, join)
	return cte
}

func (cte *Cte) Where() *ConditionBuilder {
	return cte.conditions
}

func (cte *Cte) Link(model, field string, op Operator, cteField string, trigger linkTrigger) *Cte {
	cte.liqu.linkedCte[model] = append(cte.liqu.linkedCte[model], linkedCte{
		op:       op,
		cte:      cte,
		field:    field,
		cteField: cteField,
		trigger:  trigger,
	})

	return cte
}

func (cte *Cte) Build() string {
	selectClause := "*"
	if len(cte.selects) > 0 {
		selectClause = strings.Join(cte.selects, ", ")
	}

	q := fmt.Sprintf(`SELECT %s FROM "%s"`, selectClause, cte.baseTable)
	if len(cte.joins) > 0 {
		q += " " + strings.Join(cte.joins, " ")
	}

	where := cte.liqu.registry[cte.as].branch.where.Build()
	if where != "" {
		q += " WHERE " + where
	}

	return q
}

func (cte *Cte) scan(sourceType reflect.Type, parent *branch) error {
	var (
		position    = 1
		sourceSlice = false
		sourceAs    = ""
		sourceName  = ""
	)

	// cast source to Source Type
	source, ok := (reflect.New(sourceType).Interface()).(Source)
	// if we couldn't cast the sourceType, we try to cast the first field
	if !ok {
		// first field type
		fieldType := sourceType.Field(0).Type

		// get the root element and check for slice
		fieldType, sourceSlice = getRootElement(fieldType)

		sourceName = fieldType.Name()

		// check if this fieldType is of kind Source
		var fieldCheck bool
		source, fieldCheck = ((reflect.New(fieldType)).Interface()).(Source)
		if !fieldCheck {
			return fmt.Errorf("[list] received element does not contain Source, got: %s (name %s) instead", fieldType.Kind(), fieldType.Name())
		}

		fields := cte.liqu.structFields(source)
		for k, v := range fields.fieldDatabase {
			cte.fieldDatabase[k] = v
		}

		// we are not in to main root anymore, so we need to select the result into a nested object
		sourceAs = sourceType.Field(0).Name
	}

	cte.baseTable = source.Table()

	// if the parent is nil, we are on top level,and we can populate the first branch based on what we know so far
	if parent == nil {
		structFields := cte.liqu.structFields(source)

		if sourceAs == "" {
			sourceAs = structFields.selectAs
		}

		if sourceName == "" {
			sourceName = structFields.selectAs
		}
	}

	if sourceType.NumField() >= position {
		for index := position; index < sourceType.NumField(); index++ {
			fieldType := sourceType.Field(index).Type

			if fieldType.Kind() == reflect.Slice {
				fieldType = fieldType.Elem()
			}

			var (
				source Source
				ok     bool
			)

			dbTag := sourceType.Field(index).Tag.Get("join")
			if dbTag == "" {
				dbTag = sourceType.Field(index).Name
			}

			fieldInstance := reflect.New(fieldType)
			if source, ok = (fieldInstance.Interface()).(Source); !ok {
				cte.fieldDatabase[sourceName] = dbTag

				continue
			}

			fields := cte.liqu.structFields(source)
			for k, v := range fields.fieldDatabase {
				cte.fieldDatabase[k] = v
			}

			relatedTag := sourceType.Field(index).Tag.Get("related")
			joinTag := sourceType.Field(index).Tag.Get("join")
			if joinTag == "" {
				joinTag = "LEFT"
			}

			if relatedTag != "" {
				parts := strings.Split(relatedTag, " ")

				for i := 0; i < len(parts); i++ {
					match := relatedRegex.FindStringSubmatch(parts[i])
					if len(match) != 6 {
						continue
					}

					var (
						leftTable, leftField, operator, rightTable, rightField = match[1], match[2], match[3], match[4], match[5]
					)

					condition := fmt.Sprintf("%s.%s %s %s.%s", leftTable, leftField, operator, rightTable, rightField)

					cte.Join(source.Table(), condition, joinTag)

				}
			}
		}
	}

	_ = sourceSlice

	return nil
}

func (cte *Cte) searched(s bool) *Cte {
	cte.isSearched = s

	return cte
}
