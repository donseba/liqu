package liqu

import (
	"fmt"
	"reflect"
	"strings"
)

func (l *Liqu) parseSelect(query string, reset bool) error {
	if strings.TrimSpace(query) == "" {
		return nil
	}

	selects := strings.Split(query, ",")
	for _, sel := range selects {
		parts := strings.Split(sel, ".")
		if len(parts) != 2 {
			return fmt.Errorf("invalid select format: %s", sel)
		}

		var (
			model = parts[0]
			field = parts[1]
		)

		if reset {
			l.registry[model].branch.selectedFields = make(map[string]bool)
			for _, v := range l.registry[model].branch.source.PrimaryKeys() {
				l.registry[model].branch.selectedFields[v] = true
			}
			reset = false
		}

		l.processSelect(model, field)
	}

	return nil
}

func (l *Liqu) selectsAsStruct(branch *branch) []string {
	var out []string

	for field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`"%s"."%s"`, branch.name, field))
		branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) selectsWithAlias(branch *branch) []string {
	var out []string

	for field := range branch.selectedFields {
		selectField := fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field])
		if _, ok := branch.distinctFields[field]; ok {
			selectField = fmt.Sprintf(`DISTINCT(%s)`, selectField)
		}

		out = append(out, fmt.Sprintf(`%s AS "%s"`, selectField, field))
		branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) selectsAsObjectPair(branch *branch) []string {
	var out []string

	for field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`'%s'`, field), fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
		branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) selectsWithStructAlias(branch *branch) []string {
	var (
		fields = make(map[string]bool)
		out    []string
	)

	for field := range branch.selectedFields {
		if subQ, ok := branch.subQuery[field]; ok {
			subQ.And(fmt.Sprintf(`%s.%s="%s"."%s"`, subQ.from, subQ.fieldLocal, branch.source.Table(), l.registry[branch.as].fieldDatabase[subQ.fieldParent]))
			out = append(out, fmt.Sprintf(`(%s) AS "%s"`, subQ.Build(), field))
		} else {
			out = append(out, fmt.Sprintf(`"%s"."%s" AS "%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field], field))
			branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
		}

		fields[field] = true
	}

	for field := range branch.referencedFields {
		if _, ok := fields[field]; ok {
			continue
		}

		if subQ, ok := branch.subQuery[field]; ok {
			subQ.And(fmt.Sprintf(`%s.%s="%s"."%s"`, subQ.from, subQ.fieldLocal, branch.source.Table(), l.registry[branch.as].fieldDatabase[subQ.fieldParent]))
			out = append(out, fmt.Sprintf(`(%s) AS "%s"`, subQ.Build(), field))
		} else {
			out = append(out, fmt.Sprintf(`"%s"."%s" AS "%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field], field))
			branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
		}
	}

	return out
}

func (l *Liqu) processSelect(model, field string) {
	if field == "*" {
		for field := range l.registry[model].fieldTypes {
			// if we select all fields, we don't need to select field that implements the Source interface
			if _, ok := reflect.New(l.registry[model].fieldTypes[field]).Interface().(Source); ok {
				continue
			}

			l.registry[model].branch.selectedFields[field] = true
		}
	} else {
		l.registry[model].branch.selectedFields[field] = true
	}
}
