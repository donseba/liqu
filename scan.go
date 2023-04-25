package liqu

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
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
		l.tree = &branch{
			liqu:             l,
			root:             nil,
			slice:            sourceSlice,
			anonymous:        anonymous,
			as:               sourceAs,
			name:             sourceName,
			source:           source,
			branches:         make([]*branch, 0),
			where:            NewConditionBuilder().setLiqu(l),
			selectedFields:   primaryKeys,
			referencedFields: make(map[string]bool),
		}

		// assign the parent and continue with the rest of the fields.
		parent = l.tree

		// build up the registry, so we can reference fields easier as we build up the query a bit later on
		r := &registry{
			fieldTypes:    structFields.fieldTypes,
			fieldDatabase: structFields.fieldDatabase,
			fieldSearch:   make(map[string]interface{}),
			branch:        parent,
			tableName:     source.Table(),
		}

		parent.registry = r

		l.registry[parent.as] = *r
	}

	// check the following fields if they are related
	if sourceType.NumField() >= position {
		for index := position; index < sourceType.NumField(); index++ {
			err := l.processField(sourceType.Field(index), parent)
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

func (l *Liqu) processField(structField reflect.StructField, parent *branch) error {
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
			err := l.processField(fieldType.Field(index), parent)
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
		joinTag    = strings.ToUpper(structField.Tag.Get("join"))
		relatedTag = structField.Tag.Get("related")
		//liquTag    = structField.Tag.Get("liqu")
		//dbTag      = structField.Tag.Get("db")
		//whereTag   = structField.Tag.Get("where")
		limitTag  = structField.Tag.Get("limit")
		offsetTag = structField.Tag.Get("offset")
	)

	var (
		limit  *int
		offset *int
	)

	if limitTag != "" {
		if val, ok := strconv.Atoi(limitTag); ok == nil {
			limit = &val
		}
	}

	if offsetTag != "" {
		if val, ok := strconv.Atoi(offsetTag); ok == nil {
			offset = &val
		}
	}

	structFields := l.structFields(source)
	primaryKeys := l.primaryKeys(structFields.fieldDatabase, source)
	if joinTag == "" {
		joinTag = "INNER"
	}

	currentBranch := &branch{
		liqu:             l,
		root:             parent,
		slice:            selectFieldSlice,
		as:               selectFieldAs,
		name:             selectFieldName,
		where:            NewConditionBuilder().setLiqu(l),
		limit:            limit,
		offset:           offset,
		source:           source,
		selectedFields:   primaryKeys,
		referencedFields: make(map[string]bool),
		joinDirection:    joinTag,
	}

	reg := &registry{
		fieldTypes:    structFields.fieldTypes,
		fieldDatabase: structFields.fieldDatabase,
		branch:        currentBranch,
		tableName:     source.Table(),
		fieldSearch:   make(map[string]interface{}),
	}

	currentBranch.registry = reg

	l.registry[currentBranch.as] = *reg

	parent.branches = append(parent.branches, currentBranch)

	err := l.parseRelated(relatedTag, currentBranch, parent)
	if err != nil {
		return err
	}

	err = l.scan(fieldType, currentBranch)
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

var relatedRegex = regexp.MustCompile(`([a-zA-Z]+).([a-zA-Z.]+)(=|<>|<=|>=|<|>)([a-zA-Z]+).([a-zA-Z.]+)`)

func (l *Liqu) parseRelated(tag string, branch *branch, parent *branch) error {
	relations := make([]branchRelation, 0)

	parts := strings.Split(tag, " ")
	for i := 0; i < len(parts); i++ {
		if !relatedRegex.MatchString(parts[i]) {
			continue
		}

		match := relatedRegex.FindStringSubmatch(parts[i])
		if len(match) != 6 {
			continue
		}

		var (
			leftTable, leftField, operator, rightTable, rightField = match[1], match[2], match[3], match[4], match[5]
		)

		// if the current branch is not on the left check if it is on the right and swap it if so
		if leftTable != branch.as {
			if rightTable == branch.as {
				rightTable, rightField, leftTable, leftField = leftTable, leftField, rightTable, rightField
			} else {
				return errors.New("[liqu] related expects current node to be either left or right of operator")
			}
		}

		relations = append(relations, branchRelation{
			localField:    leftField,
			operator:      operator,
			externalTable: rightTable,
			externalField: rightField,
			parent:        rightTable == parent.as,
		})

		if _, ok := l.registry[rightTable]; !ok {
			return errors.New(fmt.Sprintf("[liqu] table on the right does not exist. %s", rightTable))
		}

		// only need to expose it if the request os from a lateral join that is not within the same scope
		if rightTable != parent.as {
			Debug(rightTable)
			l.registry[rightTable].branch.referencedFields[rightField] = true
		}
	}

	branch.relations = relations
	return nil
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
