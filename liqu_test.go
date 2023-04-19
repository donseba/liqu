package liqu

import (
	"context"
	"testing"
)

type (
	Category struct {
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

	CategoryTag struct {
		CategoryID int
		TagID      int

		Tags []Tag
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
		Category Category

		CategoryTags     []CategoryTag
		CategoryProducts []CategoryProduct
	}

	Single struct {
		Category Category
	}

	SingleSlice struct {
		Category []Category
	}

	SingleAnonymous struct {
		Category
	}
)

func (m *Category) Table() string {
	return "category"
}

func (m *Category) PrimaryKeys() []string {
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

func (m *CategoryTag) Table() string {
	return "category_tag"
}

func (m *CategoryTag) PrimaryKeys() []string {
	return []string{"TagID", "categoryID"}
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

func TestSingle(t *testing.T) {
	tree := make([]Single, 0)

	li := New(context.TODO(), nil)

	err := li.FromSource(&tree)
	if err != nil {
		t.Error(err)
		return
	}

	sqlQuery, _ := li.SQL()

	expected := `SELECT count(*) OVER() AS TotalRows, to_jsonb( Category ) AS Category FROM ( SELECT category.id AS "ID" FROM category ) AS Category`
	if sqlQuery != expected {
		t.Errorf("expected %s,\ngot:%s", sqlQuery, expected)
	}
}

func TestSingleSlice(t *testing.T) {
	tree := make([]SingleSlice, 0)

	li := New(context.TODO(), nil)

	err := li.FromSource(&tree)
	if err != nil {
		t.Error(err)
		return
	}

	sqlQuery, _ := li.SQL()

	expected := `SELECT count(*) OVER() AS TotalRows, jsonb_agg( Category ) AS Category FROM ( SELECT category.id AS "ID" FROM category ) AS Category`
	if sqlQuery != expected {
		t.Errorf("expected %s,\ngot:%s", expected, sqlQuery)
	}
}

func TestSingleAnonymous(t *testing.T) {
	tree := make([]SingleAnonymous, 0)

	li := New(context.TODO(), nil)

	err := li.FromSource(&tree)
	if err != nil {
		t.Error(err)
		return
	}

	sqlQuery, _ := li.SQL()

	expected := `SELECT count(*) OVER() AS TotalRows, Category."ID" FROM ( SELECT category.id AS "ID" FROM category ) AS Category`
	if sqlQuery != expected {
		t.Errorf("expected %s,\ngot:%s", expected, sqlQuery)
	}
}
