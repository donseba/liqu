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
			l.registry[model].branch.selectedFields = make([]string, 0)
			for _, v := range l.registry[model].branch.source.PrimaryKeys() {
				l.registry[model].branch.selectedFields = appendUnique(l.registry[model].branch.selectedFields, v)
			}
			reset = false
		}

		l.processSelect(model, field)
	}

	return nil
}

func (l *Liqu) selectsAsStruct(branch *branch) []string {
	var out []string

	for _, field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`"%s"."%s"`, branch.name, field))
		branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) selectsWithAlias(branch *branch) []string {
	var out []string

	for _, field := range branch.selectedFields {
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

	for _, field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`'%s'`, field), fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
		branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) selectsWithStructAlias(branch *branch) []string {
	var out []string

	for _, field := range branch.selectedFields {
		if subQ, ok := branch.subQuery[field]; ok {
			// todo, not tested yet
			subQ.And(fmt.Sprintf(`%s.%s="%s"."%s"`, subQ.from, subQ.fieldLocal, branch.source.Table(), l.registry[branch.as].fieldDatabase[subQ.fieldParent]))

			if branch.order.HasOrderBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), branch.registry.fieldDatabase[field])) {
				out = append([]string{fmt.Sprintf(`(%s) AS "%s"`, subQ.Build(), field)}, out...)
			} else {
				out = appendUnique(out, fmt.Sprintf(`(%s) AS "%s"`, subQ.Build(), field))
			}
		} else {
			if branch.order.HasOrderBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), branch.registry.fieldDatabase[field])) {
				out = append([]string{fmt.Sprintf(`"%s"."%s" AS "%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field], field)}, out...)
			} else {
				out = appendUnique(out, fmt.Sprintf(`"%s"."%s" AS "%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field], field))
			}
			branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
		}
	}

	for field := range branch.referencedFields {
		if subQ, ok := branch.subQuery[field]; ok {
			subQ.And(fmt.Sprintf(`%s.%s="%s"."%s"`, subQ.from, subQ.fieldLocal, branch.source.Table(), l.registry[branch.as].fieldDatabase[subQ.fieldParent]))
			out = appendUnique(out, fmt.Sprintf(`(%s) AS "%s"`, subQ.Build(), field))
		} else {
			out = appendUnique(out, fmt.Sprintf(`"%s"."%s" AS "%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field], field))
			branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
		}
	}

	return out
}

func (l *Liqu) processSelect(model, field string) {
	if field == "*" {
		for f := range l.registry[model].fieldTypes {
			// if we select all fields, we don't need to select field that implements the Source interface
			if _, ok := reflect.New(l.registry[model].fieldTypes[f]).Interface().(Source); ok {
				continue
			}

			l.registry[model].branch.selectedFields = appendUnique(l.registry[model].branch.selectedFields, f)
		}
	} else {
		l.registry[model].branch.selectedFields = appendUnique(l.registry[model].branch.selectedFields, field)
	}
}
