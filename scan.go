package liqu

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

func (l *Liqu) scan(sourceType reflect.Type, parent *branch) error {
	var (
		position    = 1
		sourceSlice = false
		sourceAs    = ""
		sourceName  = ""
	)

	// cast source to Source Type
	source, ok := (reflect.New(sourceType).Interface()).(Source)

	// if we couldn't cast the sourceType we try to cast the first field
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

		// we are not in to main root anymore,so we need to select the result into a nested object
		sourceAs = sourceType.Field(0).Name
	}

	// if the parent is nil, we are on top level,and we can populate the first branch based on what we know so far
	if parent == nil {
		structFields := l.structFields(source)

		primaryKeys := l.primaryKeys(structFields.fieldDatabase, source)

		anonymous := false
		if sourceType.Field(0).Anonymous {
			anonymous = true
		}

		if sourceAs == "" {
			sourceAs = structFields.selectAs
		}

		if sourceName == "" {
			sourceName = structFields.selectAs
		}

		// set the root branch
		l.tree = branch{
			liqu:           l,
			root:           nil,
			slice:          sourceSlice,
			anonymous:      anonymous,
			As:             sourceAs,
			Name:           sourceName,
			source:         source,
			branches:       make([]*branch, 0),
			selectedFields: primaryKeys,
		}

		// assign the parent and continue with the rest of the fields.
		parent = &l.tree

		// build up the registry, so we can reference fields easier as we build up the query a bit later on
		r := &registry{
			fieldTypes:    structFields.fieldTypes,
			fieldDatabase: structFields.fieldDatabase,
			branch:        parent,
			tableName:     source.Table(),
		}

		parent.registry = r

		l.registry[parent.Name] = *r
	}

	// check the following fields if they are related
	if sourceType.NumField() >= position {
		for index := position; index < sourceType.NumField(); index++ {
			err := l.processField(sourceType.Field(index), parent /*, ""*/)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type StructFieldInfo struct {
	fieldTypes    map[string]reflect.Type
	fieldDatabase map[string]string
	selectAs      string
}

func (l *Liqu) structFields(source interface{}) StructFieldInfo {
	structFieldInfo := &StructFieldInfo{
		fieldTypes:    make(map[string]reflect.Type, 0),
		fieldDatabase: make(map[string]string, 0),
	}

	sourceElem := reflect.ValueOf(source).Elem()
	sourceType := sourceElem.Type()

	// traverse the fields of the current struct
	for i := 0; i < sourceElem.NumField(); i++ {
		// check if field is exported
		if !sourceElem.Field(i).IsValid() || !sourceElem.Field(i).CanSet() {
			continue
		}

		// get the structTags
		structTag := sourceType.Field(i).Tag

		var (
			liquTag = structTag.Get("liqu")
			dbTag   = structTag.Get("db")
		)

		if liquTag == "-" || dbTag == "-" {
			continue
		}

		// if the current field is a struct
		var hasSubField bool

		var checkSource = sourceElem.Field(i)

		if checkSource.Type().Kind() == reflect.Slice {
			checkSource = reflect.ValueOf(sourceType.Field(i).Type.Elem())
			if checkSource.Kind() == reflect.Pointer {
				checkSource = checkSource.Elem()
			}
		}

		if checkSource.Kind() == reflect.Struct {
			if (liquTag == "append" || i == 0) && sourceType.Field(i).Anonymous {
				subStructFieldInfo := l.structFields(reflect.New(sourceType.Field(i).Type).Interface())

				for k, v := range subStructFieldInfo.fieldTypes {
					structFieldInfo.fieldTypes[k] = v
				}

				for k, v := range subStructFieldInfo.fieldDatabase {
					structFieldInfo.fieldDatabase[k] = v
				}

				structFieldInfo.selectAs = sourceType.Field(i).Name

				hasSubField = true
			}
		}

		if hasSubField {
			continue
		}

		if dbTag == "" {
			dbTag = toSnakeCase(sourceType.Field(i).Name)
		}

		structFieldInfo.fieldTypes[sourceType.Field(i).Name] = sourceType.Field(i).Type
		structFieldInfo.fieldDatabase[sourceType.Field(i).Name] = dbTag
	}

	return *structFieldInfo
}

func (l *Liqu) processField(structField reflect.StructField, parent *branch /*, WrapInto string*/) error {
	fieldType := structField.Type

	if fieldType.Kind() == reflect.Slice {
		fieldType = fieldType.Elem()
	}

	fieldInstance := reflect.New(fieldType)
	if fieldIsSource, ok := (fieldInstance.Interface()).(Source); ok {
		return l.scanChild(structField, fieldIsSource, parent)
	}

	if fieldType.Kind() == reflect.Struct {
		for index := 0; index < fieldType.NumField(); index++ {
			//wrapIntoSub := structField.Name
			//if WrapInto != "" {
			//	wrapIntoSub = fmt.Sprintf("%s.%s", WrapInto, wrapIntoSub)
			//}

			err := l.processField(fieldType.Field(index), parent /*, wrapIntoSub*/)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Liqu) scanChild(structField reflect.StructField, source Source, parent *branch) error {
	fieldType := structField.Type

	var (
		selectFieldAs    = ""
		selectFieldName  = fieldType.Name()
		selectFieldSlice bool
	)

	if !structField.Anonymous {
		selectFieldAs = structField.Name
	}

	if fieldType.Kind() == reflect.Slice {
		fieldType = fieldType.Elem()
		selectFieldSlice = true
	}

	var (
	//liquTag   = structField.Tag.Get("liqu")
	//dbTag     = structField.Tag.Get("db")
	//whereTag  = structField.Tag.Get("where")
	//joinTag   = structField.Tag.Get("join")
	//limitTag  = structField.Tag.Get("limit")
	//offsetTag = structField.Tag.Get("offset")
	)

	structFields := l.structFields(source)
	primaryKeys := l.primaryKeys(structFields.fieldDatabase, source)

	currentBranch := &branch{
		liqu:           l,
		root:           parent,
		slice:          selectFieldSlice,
		As:             selectFieldAs,
		Name:           selectFieldName,
		source:         source,
		selectedFields: primaryKeys,
	}

	reg := &registry{
		fieldTypes:    structFields.fieldTypes,
		fieldDatabase: structFields.fieldDatabase,
		branch:        currentBranch,
		tableName:     source.Table(),
	}

	currentBranch.registry = reg

	l.registry[currentBranch.Name] = *reg

	parent.branches = append(parent.branches, currentBranch)

	err := l.scan(fieldType, currentBranch)
	if err != nil {
		return err
	}

	return nil
}

func (l *Liqu) primaryKeys(structFields map[string]string, source Source) []string {
	var out []string

	for _, pk := range source.PrimaryKeys() {
		for k := range structFields {
			if strings.EqualFold(k, pk) {
				out = append(out, k)
			}
		}
	}

	return out
}

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
