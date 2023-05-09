package liqu

import (
	"strings"
)

type GroupByBuilder struct {
	liqu   *Liqu
	groups []string
}

func NewGroupByBuilder() *GroupByBuilder {
	return &GroupByBuilder{
		groups: []string{},
	}
}

func (gb *GroupByBuilder) setLiqu(liqu *Liqu) *GroupByBuilder {
	gb.liqu = liqu

	return gb
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
