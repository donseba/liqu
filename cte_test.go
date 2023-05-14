package liqu

import (
	"context"
	"testing"
)

type (
	CteTagSearch struct {
		Tag        Tag
		ProjectTag ProjectTag `related:"project_tag.id_tag=tag.id"`
	}
)

func TestCte(t *testing.T) {
	filters := &Filters{
		Where: "TagSearch--Tag.Name|~~*|tagNameSearched",
	}

	def := NewDefaults().
		Where("Project.CompanyID", Equal, "0987654321")

	li := New(context.Background(), filters)

	List := make([]Single, 0)

	CteTagSearch := make([]CteTagSearch, 0)
	cte, err := NewCTE(li, CteTagSearch)
	if err != nil {
		t.Error(err)
		return
	}

	cte.Select("project_advisor.id_project")
	cte.Link("Project", "ID", In, "id_project", LinkSearch)
	li.WithCte("TagSearch", cte)
	li.WithDefaults(def)

	err = li.FromSource(List)
	if err != nil {
		t.Error(err)
		return
	}

	sqlQuery, sqlParams := li.SQL()

	t.Log(sqlQuery)
	t.Log(sqlParams)
}
