package liqu

import (
	"fmt"
	"strings"
)

type (
	CteBranchedQuery struct {
		As    string
		Query string
	}
)

func (l *Liqu) traverseCteBranch(branch *branch, parent *branch) (string, error) {
	baseCTE := newBaseQuery().setFrom(branch.registry.tableName)

	baseJoin := newBaseQuery().setFrom(branch.as)
	baseJoinConditions := NewConditionBuilder().setLiqu(l)

	var (
		selects = make([]string, 0)
	)

	branchFieldSelect := newBranchAnon()

	selects = l.selectsWithAlias(branch)

	for _, v := range branch.branches {
		err := l.traverseBranch(v, branch)
		if err != nil {
			return "", err
		}
	}

	for _, v := range branch.joinFields {
		selects = append(selects, fmt.Sprintf(`"%s"."%s" AS "%s"`, v.as, v.field, v.field))
		branch.groupBy.GroupBy(fmt.Sprintf(`"%s"."%s"`, v.as, v.field))
	}

	branchFieldSelect.setSelect(strings.Join(selects, ", "))

	for _, v := range branch.relations {
		l.registry[v.externalTable].branch.referencedFields[v.externalField] = true

		baseJoinConditions.AndRaw(fmt.Sprintf(`%s %s "%s"."%s"`,
			fmt.Sprintf(`"%s"`, v.localField),
			v.operator,
			v.externalTable,
			v.externalField,
		))
	}

	if branch.isSearched {
		branch.joinDirection = InnerJoin
		setParentJoinDirection(parent)
	}

	if cte, ok := l.linkedCte[branch.as]; ok {
		for _, v := range cte {
			l.matchCTEWithBranch(branch, v)
		}
	}

	baseCTE.setSelect(branchFieldSelect.Scrub()).setAs(branch.source.Table()).
		setJoin(strings.Join(branch.joinBranched, " ")).
		setWhere(branch.where.Build()).
		setOrderBy(branch.order.Build()).
		setGroupBy(branch.groupBy.Build())

	baseJoin.setSelect("*")
	baseJoin.setWhere(baseJoinConditions.Build())

	if branch.limit != nil {
		filters := &Filters{
			Page:    1,
			PerPage: *branch.limit,
		}

		if branch.offset != nil {
			filters.Page = *branch.offset
		}

		baseCTE.setLimit(filters)
	}

	parent.joinFields = append(parent.joinFields, branchJoinField{
		table: branch.source.Table(),
		field: branch.as,
		as:    branch.as,
		cte:   true,
		slice: branch.slice,
	})

	parent.joinBranched = append(
		parent.joinBranched,
		newLateralQuery().setQuery(baseJoin.Scrub()).setDirection(branch.joinDirection).setAs(branch.as).Scrub(),
	)

	return baseCTE.Scrub(), nil
}
