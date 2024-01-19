package liqu

import (
	"context"
	"testing"
)

type (
	Project struct {
		ID          int     `db:"id"`
		CompanyID   int     `db:"company_id"`
		Name        string  `db:"name"`
		Description string  `db:"description"`
		Volume      float64 `db:"volume"`
	}

	CategoryProject struct {
		CategoryID int
		ProjectID  int

		Project []Project
	}

	ProjectTag struct {
		ProjectID int `db:"id_project" json:"-"`
		TagID     int `db:"id_tag" json:"-"`

		Tags []Tag `related:"ProjectTags.TagID=Tags.ID" join:"left"`
	}

	Tag struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	Tree struct {
		Project Project

		ProjectTags []ProjectTag `related:"ProjectTags.ProjectID=Project.ID" join:"left"`
	}
)

func (m *Project) Table() string {
	return "project"
}

func (m *Project) PrimaryKeys() []string {
	return []string{"ID"}
}

func (m *Tag) Table() string {
	return "tag"
}

func (m *Tag) PrimaryKeys() []string {
	return []string{"ID"}
}

func (m *CategoryProject) Table() string {
	return "category_project"
}

func (m *CategoryProject) PrimaryKeys() []string {
	return []string{"CategoryID", "ProjectID"}
}

func (m *ProjectTag) Table() string {
	return "project_tag"
}

func (m *ProjectTag) PrimaryKeys() []string {
	return []string{"TagID", "ProjectID"}
}

func TestNew(t *testing.T) {
	tree := make([]Tree, 0)

	li := New(context.TODO(), nil)

	err := li.FromSource(&tree)
	if err != nil {
		t.Error(err)
		return
	}

	sql, params := li.SQL()

	expected := `SELECT coalesce(jsonb_agg(q),'[]') FROM ( SELECT count(*) OVER() AS TotalRows, to_jsonb( "Project" ) AS "Project", "ProjectTags"."ProjectTags" AS "ProjectTags" FROM ( SELECT "project"."id" AS "ID" FROM "project" GROUP BY "project"."id" ) AS "Project" LEFT JOIN LATERAL ( SELECT COALESCE(jsonb_agg( jsonb_build_object( 'TagID', "project_tag"."id_tag", 'ProjectID', "project_tag"."id_project", 'Tags', "Tags"."Tags" ) ) FILTER ( WHERE jsonb_build_object( 'TagID', "project_tag"."id_tag", 'ProjectID', "project_tag"."id_project", 'Tags', "Tags"."Tags" ) IS NOT NULL ),'[]' ) AS "ProjectTags" FROM "project_tag" LEFT JOIN LATERAL ( SELECT COALESCE(jsonb_agg( jsonb_build_object( 'ID', "tag"."id" ) ) FILTER ( WHERE jsonb_build_object( 'ID', "tag"."id" ) IS NOT NULL ),'[]' ) AS "Tags" FROM "tag" WHERE id = "project_tag"."id_tag" ) AS "Tags" ON true WHERE id_project = "Project"."ID" ) AS "ProjectTags" ON true LIMIT 25 OFFSET 0 ) q`

	if sql != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, sql)
	}

	if len(params) != 0 {
		t.Errorf("expected 0 params, got %d", len(params))
	}
}

type (
	Single struct {
		Project Project
	}

	SingleSlice struct {
		Project []Project
	}

	SingleAnonymous struct {
		Project
	}

	SingleWithCTE struct {
		Project Project

		ProjectTags []ProjectTag `liqu:"cteBranchedQueries" related:"ProjectTags.ProjectID=Project.ID" join:"left"`
	}
)

func TestWithoutJoins(t *testing.T) {
	test := []struct {
		Model    any
		Expected string
	}{
		{
			Model:    make([]Single, 0),
			Expected: `SELECT coalesce(jsonb_agg(q),'[]') FROM ( SELECT count(*) OVER() AS TotalRows, to_jsonb( "Project" ) AS "Project" FROM ( SELECT "project"."id" AS "ID" FROM "project" GROUP BY "project"."id" ) AS "Project" LIMIT 25 OFFSET 0 ) q`,
		},
		{
			Model:    make([]SingleSlice, 0),
			Expected: `SELECT coalesce(jsonb_agg(q),'[]') FROM ( SELECT count(*) OVER() AS TotalRows, COALESCE(jsonb_agg( "Project" ) FILTER ( WHERE "Project" IS NOT NULL ),'[]' ) AS "Project" FROM ( SELECT "project"."id" AS "ID" FROM "project" GROUP BY "project"."id" ) AS "Project" LIMIT 25 OFFSET 0 ) q`,
		},
		{
			Model:    make([]SingleAnonymous, 0),
			Expected: `SELECT coalesce(jsonb_agg(q),'[]') FROM ( SELECT count(*) OVER() AS TotalRows, "Project"."ID" FROM ( SELECT "project"."id" AS "ID" FROM "project" GROUP BY "project"."id" ) AS "Project" LIMIT 25 OFFSET 0 ) q`,
		},
	}

	for _, te := range test {
		filters := &Filters{
			Select: "Project.ID",
		}

		li := New(context.TODO(), filters)

		err := li.FromSource(te.Model)
		if err != nil {
			t.Error(err)
			return
		}

		sqlQuery, _ := li.SQL()

		if sqlQuery != te.Expected {
			t.Errorf("expected:\n%s,\ngot:\n%s", te.Expected, sqlQuery)
		}
	}
}

func TestWithJoins(t *testing.T) {

}

func TestWithWhere(t *testing.T) {
	filters := &Filters{
		Page:    1,
		PerPage: 25,
		OrderBy: "Project.Name|ASC,Tags.Name|DESC",
		Where:   "Project.CompanyID|=|overrideCheck,Project.Name|=|Foo",
	}

	li := New(context.TODO(), filters)

	def := NewDefaults().
		OrderBy("Project.Name", Desc).
		Where("Project.CompanyID", Equal, "11111111-0000-0000-0000-123456789012").
		Select("Project", "*").
		Select("Tags", "*")

	li.WithDefaults(def)

	tree := make([]Tree, 0)

	err := li.FromSource(tree)
	if err != nil {
		t.Error(err)
		return
	}

	sqlQuery, sqlParams := li.SQL()

	expected := `SELECT coalesce(jsonb_agg(q),'[]') FROM ( SELECT count(*) OVER() AS TotalRows, to_jsonb( "Project" ) AS "Project", "ProjectTags"."ProjectTags" AS "ProjectTags" FROM ( SELECT "project"."name" AS "Name", "project"."id" AS "ID", "project"."company_id" AS "CompanyID", "project"."description" AS "Description", "project"."volume" AS "Volume" FROM "project" WHERE "project"."company_id" = $1 AND "project"."name" = $2 GROUP BY "project"."name", "project"."id", "project"."company_id", "project"."description", "project"."volume" ORDER BY "project"."name" ASC) AS "Project" LEFT JOIN LATERAL ( SELECT COALESCE(jsonb_agg( jsonb_build_object( 'TagID', "project_tag"."id_tag", 'ProjectID', "project_tag"."id_project", 'Tags', "Tags"."Tags" ) ) FILTER ( WHERE jsonb_build_object( 'TagID', "project_tag"."id_tag", 'ProjectID', "project_tag"."id_project", 'Tags', "Tags"."Tags" ) IS NOT NULL ),'[]' ) AS "ProjectTags" FROM "project_tag" LEFT JOIN LATERAL ( SELECT COALESCE(jsonb_agg( jsonb_build_object( 'ID', "tag"."id", 'Name', "tag"."name" ) ORDER BY "tag"."name" DESC ) FILTER ( WHERE jsonb_build_object( 'ID', "tag"."id", 'Name', "tag"."name" ) IS NOT NULL ),'[]' ) AS "Tags" FROM "tag" WHERE id = "project_tag"."id_tag" ) AS "Tags" ON true WHERE id_project = "Project"."ID" ) AS "ProjectTags" ON true ORDER BY "Name" ASC LIMIT 25 OFFSET 0 ) q`
	if sqlQuery != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, sqlQuery)
	}

	if len(sqlParams) != 2 {
		t.Errorf("expected 2 params, got %d", len(sqlParams))
	}
}

func TestWithSubQuery(t *testing.T) {
	filters := &Filters{
		Page:    1,
		PerPage: 25,
		OrderBy: "Project.Name|ASC",
		Select:  "Project.*",
	}

	volumeSQ := NewSubQuery("Project", "Volume").
		Relate("id_project", "ID").
		Select("SUM(volume)").
		From("project_time_entry")

	li := New(context.TODO(), filters).
		WithSubQuery(volumeSQ)

	tree := make([]Single, 0)

	err := li.FromSource(tree)
	if err != nil {
		t.Error(err)
		return
	}

	sqlQuery, sqlParams := li.SQL()

	expected := `SELECT coalesce(jsonb_agg(q),'[]') FROM ( SELECT count(*) OVER() AS TotalRows, to_jsonb( "Project" ) AS "Project" FROM ( SELECT "project"."name" AS "Name", "project"."id" AS "ID", "project"."company_id" AS "CompanyID", "project"."description" AS "Description", (SELECT SUM(volume) FROM "project_time_entry" WHERE project_time_entry.id_project="project"."id") AS "Volume" FROM "project" GROUP BY "project"."name", "project"."id", "project"."company_id", "project"."description" ORDER BY "project"."name" ASC) AS "Project" ORDER BY "Name" ASC LIMIT 25 OFFSET 0 ) q`

	if len(sqlQuery) != len(expected) {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, sqlQuery)
	}
	// still need to ensure the order of the fields is the same
	t.Log(sqlQuery)

	if len(sqlParams) != 0 {
		t.Errorf("expected 0 params, got %d", len(sqlParams))
	}
}

func TestWithCTE(t *testing.T) {
	filters := &Filters{
		Page:    1,
		PerPage: 25,
		OrderBy: "Project.Name|ASC",
		Select:  "Project.Name",
	}

	def := NewDefaults().
		OrderBy("Project.Name", Desc).
		//Where("Project.CompanyID", Equal, "11111111-0000-0000-0000-000000000000").
		//Where("ProjectTags.TagID", Equal, "12345678-0000-0000-0000-FAKETAGID000").
		Select("Project", "*").
		Select("Tags", "*")

	li := New(context.TODO(), filters).WithDefaults(def)

	tree := make([]SingleWithCTE, 0)

	err := li.FromSource(tree)
	if err != nil {
		t.Error(err)
		return
	}

	sqlQuery, sqlParams := li.SQL()

	expected := `SELECT coalesce(jsonb_agg(q),'[]') FROM ( SELECT count(*) OVER() AS TotalRows, to_jsonb( "Project" ) AS "Project", "ProjectTags"."ProjectTags" AS "ProjectTags" FROM ( SELECT "project"."name" AS "Name", "project"."id" AS "ID" FROM "project" GROUP BY "project"."name", "project"."id" ORDER BY "project"."name" ASC) AS "Project" LEFT JOIN LATERAL ( SELECT COALESCE(jsonb_agg( jsonb_build_object( 'TagID', "project_tag"."id_tag", 'ProjectID', "project_tag"."id_project", 'Tags', "Tags"."Tags" ) ) FILTER ( WHERE jsonb_build_object( 'TagID', "project_tag"."id_tag", 'ProjectID', "project_tag"."id_project", 'Tags', "Tags"."Tags" ) IS NOT NULL ),'[]' ) AS "ProjectTags" FROM "project_tag" LEFT JOIN LATERAL ( SELECT COALESCE(jsonb_agg( jsonb_build_object( 'ID', "tag"."id", 'Name', "tag"."name" ) ) FILTER ( WHERE jsonb_build_object( 'ID', "tag"."id", 'Name', "tag"."name" ) IS NOT NULL ),'[]' ) AS "Tags" FROM "tag" WHERE id = "project_tag"."id_tag" ) AS "Tags" ON true WHERE id_project = "Project"."ID" ) AS "ProjectTags" ON true ORDER BY "Name" ASC LIMIT 25 OFFSET 0 ) q`

	if sqlQuery != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, sqlQuery)
	}

	if len(sqlParams) != 0 {
		t.Errorf("expected 0 params, got %d", len(sqlParams))
	}
}
