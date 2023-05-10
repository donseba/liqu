package liqu

type (
	branch struct {
		liqu             *Liqu
		root             *branch
		parent           *branch
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
		selectedFields   map[string]bool
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
	}

	branchRelation struct {
		localField    string
		operator      string
		externalTable string
		externalField string
		parent        bool
	}

	joinOperator string
)

const (
	leftJoin  joinOperator = "LEFT"
	InnerJoin              = "INNER"
	fullJoin               = "FULL"
)
