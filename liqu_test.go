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

	CategoryProduct struct {
		CategoryID int
		productID  int

		Products []Product
	}

	Product struct {
		ID   int    `db:"id"`
		Name string `db:"name"`

		ProductTags []ProductTag
	}

	ProjectTag struct {
		ProjectID int `db:"id_project" json:"-"`
		TagID     int `db:"id_tag" json:"-"`

		Tags []Tag `related:"ProjectTags.TagID=Tags.ID" join:"left"`
	}

	ProductTag struct {
		productID int
		TagID     int

		Tags []Tag
	}

	Tag struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	Tree struct {
		Project Project

		ProjectTags []ProjectTag `related:"ProjectTags.ProjectID=Project.ID" join:"left"`
		//ProjectProducts []CategoryProduct `related:"ProjectProducts.ProjectID=Project.ID"`
	}
)

func (m *Project) Table() string {
	return "project"
}

func (m *Project) PrimaryKeys() []string {
	return []string{"ID"}
}

func (m *Product) Table() string {
	return "product"
}

func (m *Product) PrimaryKeys() []string {
	return []string{"ID"}
}

func (m *Tag) Table() string {
	return "tag"
}

func (m *Tag) PrimaryKeys() []string {
	return []string{"ID"}
}

func (m *CategoryProduct) Table() string {
	return "category_product"
}

func (m *CategoryProduct) PrimaryKeys() []string {
	return []string{"CategoryID", "productID"}
}

func (m *ProductTag) Table() string {
	return "product_tag"
}

func (m *ProductTag) PrimaryKeys() []string {
	return []string{"TagID", "productID"}
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

	query, params := li.SQL()
	t.Log(query)
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
			Expected: `SELECT count(*) OVER() AS TotalRows, to_jsonb( Category ) AS Category FROM ( SELECT category.id AS "ID" FROM category ) AS Category`,
		},
		{
			Model:    make([]SingleSlice, 0),
			Expected: `SELECT count(*) OVER() AS TotalRows, jsonb_agg( Category ) AS Category FROM ( SELECT category.id AS "ID" FROM category ) AS Category`,
		},
		{
			Model:    make([]SingleAnonymous, 0),
			Expected: `SELECT count(*) OVER() AS TotalRows, Category."ID" FROM ( SELECT category.id AS "ID" FROM category ) AS Category`,
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
			t.Errorf("expected %s,\ngot:%s", te.Expected, sqlQuery)
		}
	}
}
