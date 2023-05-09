package liqu

import (
	"fmt"
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
			reset = false
		}

		l.processSelect(model, field)
	}

	return nil
}

func (l *Liqu) selectsAsStruct(branch *branch) []string {
	var out []string

	for field, _ := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`%s."%s"`, branch.name, field))
		branch.groupBy.GroupBy(fmt.Sprintf(`%s.%s`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) selectsAsObjectPair(branch *branch) []string {
	var out []string

	for field, _ := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`'%s'`, field), fmt.Sprintf(`%s.%s`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
		branch.groupBy.GroupBy(fmt.Sprintf(`%s.%s`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) selectsWithStructAlias(branch *branch) []string {
	var out []string

	for field, _ := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`%s.%s AS "%s"`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field], field))
		branch.groupBy.GroupBy(fmt.Sprintf(`%s.%s`, branch.source.Table(), l.registry[branch.as].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) processSelect(model, field string) {
	if field == "*" {
		for field, _ := range l.registry[model].fieldTypes {
			l.registry[model].branch.selectedFields[field] = true
		}
	} else {
		l.registry[model].branch.selectedFields[field] = true
	}
}
