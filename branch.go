package liqu

type (
	branch struct {
		liqu             *Liqu
		root             *branch
		slice            bool
		anonymous        bool
		as               string
		name             string
		where            *ConditionBuilder
		source           Source
		limit            *int
		offset           *int
		registry         *registry
		branches         []*branch
		relations        []branchRelation
		selectedFields   []string
		referencedFields map[string]bool

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
)
