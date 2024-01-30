package liqu

import (
	"fmt"
	"reflect"
	"strings"
)

type Aggregator string

const (
	AggCount Aggregator = "COUNT"
	AggSum   Aggregator = "SUM"
	AggAvg   Aggregator = "AVG"
	AggMin   Aggregator = "MIN"
	AggMax   Aggregator = "MAX"
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
		out = append(out, fmt.Sprintf(`"%s"."%s" AS "%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field], field))

		if len(branch.aggregateFields) > 0 {
			branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
		}
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

func (l *Liqu) aggregateWithStructAlias(branch *branch) []string {
	var out []string

	for _, field := range branch.aggregateFields {
		out = append(out, fmt.Sprintf(`%s("%s"."%s") AS "%s"`, field.Func, branch.source.Table(), l.registry[branch.as].fieldDatabase[field.Field], field.Alias))
	}

	return out
}

func (l *Liqu) aggregateWithAlias(branch *branch) []string {
	var out []string

	for _, field := range branch.aggregateFields {
		if branch.as == l.tree.as && l.tree.anonymous {
			out = append(out, fmt.Sprintf(`%s("%s"."%s") AS "%s"`, field.Func, branch.source.Table(), l.registry[branch.as].fieldDatabase[field.Field], field.Alias))
		} else {
			out = append(out, fmt.Sprintf(`%s("%s"."%s") AS "%s"`, field.Func, branch.as, field.Field, field.Alias))
		}
	}

	return out
}

func (l *Liqu) processSelectAggregate(model, field, alias string, funC Aggregator) {
	l.registry[model].branch.aggregateFields = append(l.registry[model].branch.aggregateFields, aggregateField{
		Func:  funC,
		Field: field,
		Alias: alias,
	})
}
