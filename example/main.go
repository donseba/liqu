package main

import (
	"context"
	"github.com/donseba/liqu"
	"log"
)

type (
	Article struct {
		ID         int    `db:"id"`
		Title      string `db:"title"`
		Content    string `db:"content"`
		AuthorID   int    `db:"author_id"`
		CategoryID int    `db:"category_id"`
	}

	Author struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	Category struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	ArticleList struct {
		Article  Article
		Author   Author   `join:"right" related:"Author.ID=Article.AuthorID"`
		Category Category `join:"left" related:"Category.ID=Article.CategoryID"`
	}
)

func (*Article) Table() string {
	return "article"
}

func (*Article) PrimaryKeys() []string {
	return []string{"ID"}
}

func (m *Author) Table() string {
	return "author"
}

func (m *Author) PrimaryKeys() []string {
	return []string{"ID"}
}

func (m *Category) Table() string {
	return "author"
}

func (m *Category) PrimaryKeys() []string {
	return []string{"ID"}
}

func main() {
	list := make([]ArticleList, 0)

	li := liqu.New(context.TODO(), nil)

	err := li.FromSource(&list)
	if err != nil {
		log.Fatal(err)
		return
	}

	sqlQuery, sqlParams := li.SQL()
	liqu.Debug(sqlQuery, sqlParams)

	var result string
	// now you can pass the sqlQuery and sqlParams to your favorite sql executor.
	// result = sql.SelectString(sqlQuery, sqlParams...)

	result = li.PostProcess(result)

	// after the PostProcess you kan fetch the paging params from the filters
	filters := li.Filters()
	liqu.Debug(filters)
}
