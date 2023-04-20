package liqu

type (
	branch struct {
		liqu             *Liqu
		root             *branch
		slice            bool
		anonymous        bool
		As               string
		Name             string
		source           Source
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
		Table string
		Field string
		As    string
	}

	branchOptions struct {
		Where string
	}

	branchRelation struct {
		localField    string
		operator      string
		externalTable string
		externalField string
		parent        bool
	}
)
