package liqu

type (
	branch struct {
		liqu             *Liqu
		root             *branch
		parent           *branch
		isCTE            bool
		slice            bool
		anonymous        bool
		as               string
		name             string
		where            *ConditionBuilder
		isSearched       bool
		order            *OrderBuilder
		groupBy          *GroupByBuilder
		source           Source
		limit            *int
		offset           *int
		registry         *registry
		branches         []*branch
		relations        []branchRelation
		selectedFields   []string
		distinctFields   map[string]bool
		referencedFields map[string]bool
		subQuery         map[string]*SubQuery

		joinDirection string
		joinFields    []branchJoinField
		joinBranched  []string
	}

	branchJoinField struct {
		table string
		field string
		as    string
		cte   bool
		slice bool
	}

	branchRelation struct {
		localField    string
		operator      string
		externalTable string
		externalField string
		parent        bool
	}

	linkedCte struct {
		op       Operator
		cte      *Cte
		field    string
		cteField string
		trigger  linkTrigger
	}

	joinOperator string
)

const (
	leftJoin  joinOperator = "LEFT"
	InnerJoin              = "INNER"
	fullJoin               = "FULL"
)
