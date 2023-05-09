package liqu

import (
	"fmt"
	"strings"
)

type GroupByBuilder struct {
	groups []string
}

func NewGroupByBuilder() *GroupByBuilder {
	return &GroupByBuilder{
		groups: []string{},
	}
}

func (gb *GroupByBuilder) GroupBy(column string) *GroupByBuilder {
	for _, v := range gb.groups {
		if v == column {
			return gb
		}
	}

	gb.groups = append(gb.groups, column)
	return gb
}

func (gb *GroupByBuilder) Build() string {
	return strings.Join(gb.groups, ", ")
}

func (l *Liqu) parseGroupBy(query string) error {
	if strings.TrimSpace(query) == "" {
		return nil
	}

	groupBys := strings.Split(query, ",")

	for _, groupby := range groupBys {

		err := l.processGroupBy(groupby)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Liqu) processGroupBy(col string) error {
	var (
		model  string
		field  string
		column string
	)

	if strings.Contains(col, ".") {
		el := strings.Split(col, ".")
		model = el[0]
		field = el[1]
	} else {
		model = l.tree.as
		field = col
	}

	var ok bool
	if column, ok = l.registry[model].fieldDatabase[field]; !ok {
		return fmt.Errorf("invalid group by field %s", col)
	}

	column = fmt.Sprintf(`"%s"."%s"`, l.registry[model].tableName, column)

	l.registry[model].branch.groupBy.GroupBy(column)

	return nil
}
