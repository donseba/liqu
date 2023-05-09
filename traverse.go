package liqu

import (
	"fmt"
	"strings"
)

func (l *Liqu) traverse() error {
	root := newRootQuery()

	if l.sourceSlice {
		root.SetTotalRows("count(*) OVER() AS TotalRows,")
	}

	rootFieldSelect := newBranchAnon()
	if !l.tree.anonymous {
		if l.tree.slice {
			rootFieldSelect = newBranchSlice()
		} else {
			rootFieldSelect = newBranchSingle()
		}
		rootFieldSelect.setSelect(l.tree.as).setAs(l.tree.as)
	} else {
		rootFieldSelect.setSelect(strings.Join(l.selectsAsStruct(l.tree), ", "))
	}

	for _, v := range l.tree.branches {
		err := l.traverseBranch(v, l.tree)
		if err != nil {
			return err
		}
	}

	root.setJoin(strings.Join(l.tree.joinBranched, " "))

	base := newBaseQuery().
		setFrom(l.tree.registry.tableName).
		setSelect(strings.Join(l.selectsWithStructAlias(l.tree), ", "))

	rootSelects := []string{rootFieldSelect.Scrub()}

	for _, v := range l.tree.joinFields {
		rootSelects = append(rootSelects, fmt.Sprintf(`%s.%s AS "%s"`, v.as, v.field, v.field))
	}

	root.setSelect(strings.Join(rootSelects, ", ")).
		setFrom(base.Scrub()).
		setAs(l.tree.as).
		setLimit(l.filters).
		setWhere(l.tree.where.Build()).
		setOrderBy(l.tree.order.Build()).
		setGroupBy(l.tree.groupBy.Build())

	var wrapper *query
	if l.sourceSlice {
		wrapper = newSliceQuery()
	} else {
		wrapper = newSingleQuery()
	}

	l.sqlQuery = wrapper.setQuery(root.Scrub()).Scrub()

	return nil
}

func (l *Liqu) traverseBranch(branch *branch, parent *branch) error {
	if len(branch.relations) == 0 {
		return nil
	}

	base := newBaseQuery().setFrom(branch.registry.tableName)

	var (
		selects = make([]string, 0)
	)

	branchFieldSelect := newBranchAnon()
	if !branch.anonymous {
		if branch.slice {
			branchFieldSelect = newBranchSlice()
		} else {
			branchFieldSelect = newBranchSingle()
		}
		selects = l.selectsAsObjectPair(branch)
	}

	parent.joinFields = append(parent.joinFields, branchJoinField{
		table: branch.source.Table(),
		field: branch.as,
		as:    branch.as,
	})

	for _, v := range branch.branches {
		err := l.traverseBranch(v, branch)
		if err != nil {
			return err
		}
	}

	for _, v := range branch.joinFields {
		selects = append(selects, fmt.Sprintf(`'%s', %s.%s`, v.field, v.as, v.field))
		//branch.groupBy.GroupBy(v.field)
	}

	branchFieldSelect.setSelect(fmt.Sprintf("jsonb_build_object( %s )", strings.Join(selects, ", "))).setAs(branch.as)

	for _, v := range branch.relations {
		externalField := l.registry[v.externalTable].fieldDatabase[v.externalField]
		externalTable := l.registry[v.externalTable].tableName
		if v.parent {
			if l.tree.as == v.externalTable {
				externalField = fmt.Sprintf(`"%s"`, v.externalField)
			}
		} else {
			externalTable = v.externalTable
		}

		branch.where.AndRaw(fmt.Sprintf("%s %s %s.%s",
			l.registry[branch.as].fieldDatabase[v.localField],
			v.operator,
			externalTable,
			externalField,
		))
	}

	selectsWithReferences := make([]string, 0)
	for k, _ := range branch.referencedFields {
		selectsWithReferences = append(selectsWithReferences, branch.registry.fieldDatabase[k])
		//branch.groupBy.GroupBy(branch.registry.fieldDatabase[k])
	}

	selectsWithReferences = append(selectsWithReferences, branchFieldSelect.Scrub())

	base.setSelect(strings.Join(selectsWithReferences, ", ")).
		setJoin(strings.Join(branch.joinBranched, " ")).
		setWhere(branch.where.Build()).
		setOrderBy(branch.order.Build())
	//setGroupBy(branch.groupBy.Build())

	if branch.limit != nil {
		paging := &Filters{
			PerPage: *branch.limit,
		}
		if branch.offset == nil {
			paging.Page = *branch.offset
		}
		base.setLimit(paging)
	}

	parent.joinBranched = append(
		parent.joinBranched,
		newLateralQuery().setQuery(base.Scrub()).setDirection(branch.joinDirection).setAs(branch.as).Scrub(),
	)

	return nil
}
