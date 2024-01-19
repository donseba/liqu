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

	expected := `WITH "TagSearch" AS ( SELECT project_advisor.id_project FROM "tag" LEFT JOIN project_tag ON project_tag.id_tag = tag.id WHERE "tag"."name" ~~* $2 ) SELECT coalesce(jsonb_agg(q),'[]') FROM ( SELECT count(*) OVER() AS TotalRows, to_jsonb( "Project" ) AS "Project" FROM ( SELECT "project"."id" AS "ID", "project"."company_id" AS "CompanyID" FROM "project" WHERE "project"."company_id" = $1 AND "project"."id" IN (SELECT * FROM "TagSearch") GROUP BY "project"."id", "project"."company_id" ) AS "Project" LIMIT 25 OFFSET 0 ) q`

	if sqlQuery != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, sqlQuery)
	}

	if len(sqlParams) != 2 {
		t.Errorf("expected 2 params, got %d", len(sqlParams))
	}

	if sqlParams[0] != "0987654321" {
		t.Errorf("expected 0987654321, got %s", sqlParams[0])
	}

	if sqlParams[1] != "%tagNameSearched%" {
		t.Errorf("expected %%tagNameSearched%%, got %s", sqlParams[1])
	}
}
