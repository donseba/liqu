package liqu

import (
	"context"
	"testing"
)

type (
	Project struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
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

		ProjectTags []ProjectTag `related:"ProjectTags.ProjectID=Project.ID" join:"left" limit:"1" offset:"0"`
		//ProjectCategories []CategoryProject `related:"ProjectCategories.ProjectID=Project.ID"`
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
	t.Log(sql)
	t.Log(params)
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
)

func TestWithoutJoins(t *testing.T) {
	test := []struct {
		Model    any
		Expected string
	}{
		{
			Model:    make([]Single, 0),
			Expected: `SELECT count(*) OVER() AS TotalRows, to_jsonb( Project ) AS Project FROM ( SELECT project.id AS "ID" FROM project ) AS Project LIMIT 25 OFFSET 0`,
		},
		{
			Model:    make([]SingleSlice, 0),
			Expected: `SELECT count(*) OVER() AS TotalRows, coalesce(jsonb_agg( Project ), '[]') AS Project FROM ( SELECT project.id AS "ID" FROM project ) AS Project LIMIT 25 OFFSET 0`,
		},
		{
			Model:    make([]SingleAnonymous, 0),
			Expected: `SELECT count(*) OVER() AS TotalRows, Project."ID" FROM ( SELECT project.id AS "ID" FROM project ) AS Project LIMIT 25 OFFSET 0`,
		},
	}

	for _, te := range test {
		li := New(context.TODO(), nil)

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
		Where:   "Project.Name|ILIKE|John,Tags.Name|=|tagName",
	}

	li := New(context.TODO(), filters)

	tree := make([]Tree, 0)

	err := li.FromSource(tree)
	if err != nil {
		t.Error(err)
		return
	}

	sqlQuery, sqlParams := li.SQL()

	t.Log(sqlQuery)
	t.Log(sqlParams)
}
