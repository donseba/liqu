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
		rootFieldSelect.setSelect(fmt.Sprintf(`"%s"`, l.tree.as)).setAs(l.tree.as)
	} else {
		rootFieldSelect.setSelect(strings.Join(l.selectsAsStruct(l.tree), ", "))
	}

	whereNulls := NewConditionBuilder()
	for _, v := range l.tree.branches {
		// if it is a Cte, we branch of and threat it as a root element with no parent.
		if v.isCTE {
			cte, err := l.traverseCteBranch(v, l.tree)
			if err != nil {
				return err
			}
			l.cteBranchedQueries = append(l.cteBranchedQueries, &CteBranchedQuery{
				As:    v.as,
				Query: cte,
			})
		} else {
			err := l.traverseBranch(v, l.tree)
			if err != nil {
				return err
			}
		}

		if v.joinDirection == InnerJoin && !v.isCTE {
			whereNulls.AndIsNotNull(fmt.Sprintf(`"%s"`, v.as))
		}
	}

	root.setJoin(strings.Join(l.tree.joinBranched, " "))

	base := newBaseQuery().
		setFrom(l.tree.registry.tableName).
		setSelect(strings.Join(l.selectsWithStructAlias(l.tree), ", "))

	rootSelects := []string{rootFieldSelect.Scrub()}

	cteGroupBy := NewGroupByBuilder()

	var hasSubCTE bool
	for _, v := range l.tree.joinFields {
		if v.cte {
			subCTE := newBranchSingle().setAs(fmt.Sprintf(`"%s"`, v.field))
			if v.slice {
				subCTE = newBranchCteSlice().setAs(v.field)
				cteGroupBy.GroupBy(fmt.Sprintf(`"%s"`, l.tree.as))
				cteGroupBy.GroupBy(fmt.Sprintf(`"%s"."Result"`, v.field))
				rootSelects = append(rootSelects, subCTE.setSelect(fmt.Sprintf(`"%s"."Result"`, v.field)).Scrub())
			} else {
				rootSelects = append(rootSelects, subCTE.setSelect(fmt.Sprintf(`"%s"`, v.field)).Scrub())
			}
			hasSubCTE = true
		} else {
			rootSelects = append(rootSelects, fmt.Sprintf(`"%s"."%s" AS "%s"`, v.as, v.field, v.field))
		}
	}

	if cte, ok := l.linkedCte[l.tree.as]; ok {
		for _, v := range cte {
			l.matchCTEWithBranch(l.tree, v)
		}
	}

	if hasSubCTE {
		for _, v := range l.tree.branches {
			if v.isCTE {
				continue
			}
			cteGroupBy.GroupBy(fmt.Sprintf(`"%s"`, v.as))
		}
	}

	root.setSelect(strings.Join(rootSelects, ", ")).
		setFrom(base.Scrub()).
		setAs(l.tree.as).
		setLimit(l.filters).
		setWhere(l.tree.where.Build()).
		setWhereNulls(whereNulls.Build()).
		setOrderBy(l.tree.order.Build())

	if l.tree.order.Build() != "" {
		eo, err := ExtractOrders(l.tree.order.Build())
		if err == nil {
			no := NewOrderBuilder()
			for _, v := range eo {
				if l.tree.registry.tableName != v.Table {
					continue
				}

				for k, fd := range l.tree.registry.fieldDatabase {
					if fd == v.Column {
						if len(cteGroupBy.groups) > 0 {
							cteGroupBy.GroupBy(fmt.Sprintf(`"%s"`, k))
						}
						no.OrderBy(fmt.Sprintf(`"%s"`, k), OrderDirection(v.Direction))
					}
				}
			}

			root.setOrderByParent(no.Build())
		}
	}

	root.setGroupBy(l.tree.groupBy.Build()).
		setGroupByCTE(cteGroupBy.Build())

	var wrapper *query
	if l.sourceSlice {
		wrapper = newSliceQuery()
	} else {
		wrapper = newSingleQuery()
	}

	var cteQueries []string
	for _, v := range l.cteBranchedQueries {
		cteQueries = append(cteQueries, newCteQuery().setWith(v.As).setQuery(v.Query).Scrub())
	}

	for as, cte := range l.cte {
		cteQueries = append(cteQueries, newCteQuery().setWith(as).setQuery(cte.Build()).Scrub())
	}

	l.sqlQuery = wrapper.setCTE(strings.Join(cteQueries, ", ")).setQuery(root.Scrub()).Scrub()

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
		selects = append(selects, fmt.Sprintf(`'%s', "%s"."%s"`, v.field, v.as, v.field))
	}

	branchFieldSelect.setSelect(fmt.Sprintf("jsonb_build_object( %s )", strings.Join(selects, ", "))).setAs(branch.as)
	// set the order_by inside the branch just behind the jsonb_agg so we can order on the fields inside the jsonb_agg
	branchOrder := branch.order.Build()
	branchFieldSelect.setOrderBy(branchOrder)

	for _, v := range branch.relations {
		externalField := l.registry[v.externalTable].fieldDatabase[v.externalField]
		externalTable := l.registry[v.externalTable].tableName
		if v.parent {
			if l.tree.as == v.externalTable {
				externalField = fmt.Sprintf(`%s`, v.externalField)
				externalTable = v.externalTable
			}
		} else {
			externalTable = v.externalTable
		}

		branch.where.AndRaw(fmt.Sprintf(`%s %s "%s"."%s"`,
			l.registry[branch.as].fieldDatabase[v.localField],
			v.operator,
			externalTable,
			externalField,
		))
	}

	selectsWithReferences := make([]string, 0)
	for k := range branch.referencedFields {
		selectsWithReferences = append(selectsWithReferences, branch.registry.fieldDatabase[k])
	}

	selectsWithReferences = append(selectsWithReferences, branchFieldSelect.Scrub())

	if branch.isSearched {
		branch.joinDirection = InnerJoin
		setParentJoinDirection(parent)
	}

	if cte, ok := l.linkedCte[branch.as]; ok {
		for _, v := range cte {
			l.matchCTEWithBranch(branch, v)
		}
	}

	base.setSelect(strings.Join(selectsWithReferences, ", ")).
		setJoin(strings.Join(branch.joinBranched, " ")).
		setWhere(branch.where.Build())

	if branch.parent.isCTE {
		base.setGroupBy(branch.groupBy.Build())
	}

	if branch.limit != nil {
		filters := &Filters{
			Page:    1,
			PerPage: *branch.limit,
		}

		if branch.offset != nil {
			filters.Page = *branch.offset
		}

		base.setLimit(filters)
	}

	parent.joinBranched = append(
		parent.joinBranched,
		newLateralQuery().setQuery(base.Scrub()).setDirection(branch.joinDirection).setAs(branch.as).Scrub(),
	)

	return nil
}

// if we start to search within a relation, we want to make sure we only return the rows having these results.
func setParentJoinDirection(parent *branch) {
	if parent == nil {
		return
	}

	parent.joinDirection = InnerJoin

	if parent.parent != nil {
		setParentJoinDirection(parent.parent)
	}
}

func (l *Liqu) matchCTEWithBranch(branch *branch, linkedCte linkedCte) {
	switch linkedCte.trigger {
	case LinkSearch:
		if linkedCte.cte.isSearched {
			l.applyCTEonBranch(branch, linkedCte)
		}
	case LinkAlways:
		l.applyCTEonBranch(branch, linkedCte)
	}
}

func (l *Liqu) applyCTEonBranch(branch *branch, linkedCte linkedCte) {
	baseQuery := newBaseQuery()

	switch linkedCte.op {
	case In:
		baseQuery.setFrom(linkedCte.cte.as).setSelect("*")
		branch.where.AndRaw(fmt.Sprintf(`"%s"."%s" IN (%s)`, branch.source.Table(), branch.registry.fieldDatabase[linkedCte.field], baseQuery.Scrub()))
	}
}
