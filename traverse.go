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

	if !l.tree.anonymous {
		rootSelect := fmt.Sprintf("to_jsonb( :select: ) AS %s", l.tree.As)
		if l.tree.slice {
			rootSelect = fmt.Sprintf("jsonb_agg( :select: ) AS %s", l.tree.As)
		}
		root.setSelect(rootSelect)
	}

	for _, v := range l.tree.branches {
		err := l.traverseBranch(v, l.tree)
		if err != nil {
			return err
		}
	}

	root.setJoin(strings.Join(l.tree.joinBranched, " "))

	base := newBaseQuery()
	base.setFrom(l.tree.registry.tableName)

	base.setSelect(strings.Join(l.selectsWithStructAlias(l.tree), ","))

	var rootSelects []string
	if l.tree.anonymous {
		rootSelects = l.selectsAsStruct(l.tree)
	} else {
		rootSelects = []string{l.tree.As}
	}

	for _, v := range l.tree.joinFields {
		rootSelects = append(rootSelects, fmt.Sprintf(`%s.%s AS "%s"`, v.As, v.Field, v.Field))
	}

	root.setSelect(strings.Join(rootSelects, ",")).
		setFrom(base.Scrub()).
		setAs(l.tree.As, l.tree.registry.tableName)

	l.sqlQuery = root.Scrub()

	return nil
}

func (l *Liqu) traverseBranch(branch *branch, parent *branch) error {
	if len(branch.relations) == 0 {
		return nil
	}

	base := newBaseQuery().setFrom(branch.registry.tableName)

	var (
		selects = make([]string, 0)
		wheres  = make([]string, 0)
		groupBy = make([]string, 0)
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
		Table: branch.source.Table(),
		Field: branch.As,
		As:    branch.As,
	})

	for _, v := range branch.branches {
		err := l.traverseBranch(v, branch)
		if err != nil {
			return err
		}
	}

	for _, v := range branch.joinFields {
		selects = append(selects, fmt.Sprintf(`'%s', %s.%s`, v.Field, v.As, v.Field))
	}

	branchFieldSelect.setSelect(fmt.Sprintf("jsonb_build_object( %s )", strings.Join(selects, ", "))).setAs(branch.As, branch.Name)

	for _, v := range branch.relations {
		externalField := l.registry[v.externalTable].fieldDatabase[v.externalField]
		externalTable := l.registry[v.externalTable].tableName
		if v.parent {
			if l.tree.As == v.externalTable {
				externalField = fmt.Sprintf(`"%s"`, v.externalField)
			}
		} else {
			externalTable = v.externalTable
		}

		wheres = append(wheres,
			fmt.Sprintf("%s %s %s.%s",
				l.registry[branch.As].fieldDatabase[v.localField],
				v.operator,
				externalTable,
				externalField,
			),
		)
	}

	selectsWithReferences := []string{}
	for k, _ := range branch.referencedFields {
		selectsWithReferences = append(selectsWithReferences, branch.registry.fieldDatabase[k])
		groupBy = append(groupBy, branch.registry.fieldDatabase[k])
	}

	selectsWithReferences = append(selectsWithReferences, branchFieldSelect.Scrub())

	base.setSelect(strings.Join(selectsWithReferences, ","))
	base.setJoin(strings.Join(branch.joinBranched, " "))
	base.setWhere(strings.Join(wheres, " AND "))
	base.setGroupBy(groupBy)

	parent.joinBranched = append(
		parent.joinBranched,
		newLateralQuery().setQuery(base.Scrub()).setDirection(branch.joinDirection).setAs(branch.As, branch.Name).Scrub(),
	)

	return nil
}

func (l *Liqu) selectsAsStruct(branch *branch) []string {
	var out []string

	for _, field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`%s."%s"`, branch.Name, field))
	}

	return out
}

func (l *Liqu) selectsAsObjectPair(branch *branch) []string {
	var out []string

	for _, field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`'%s'`, field), fmt.Sprintf(`%s.%s`, branch.source.Table(), l.registry[branch.As].fieldDatabase[field]))
	}

	return out
}

func (l *Liqu) selectsWithStructAlias(branch *branch) []string {
	var out []string

	for _, field := range branch.selectedFields {
		out = append(out, fmt.Sprintf(`%s.%s AS "%s"`, branch.source.Table(), l.registry[branch.As].fieldDatabase[field], field))
	}

	return out
}
