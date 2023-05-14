package main

import (
	"context"
	"encoding/json"
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
		Article    Article
		Author     Author     `join:"right" related:"Author.ID=Article.AuthorID"`
		Categories []Category `liqu:"cte" join:"left" related:"Categories.ID=Article.CategoryID"`
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
	return "category"
}

func (m *Category) PrimaryKeys() []string {
	return []string{"ID"}
}

func main() {
	// initiate a slice of the object to process
	list := make([]ArticleList, 0)

	// define defaults to be used for the query
	def := liqu.NewDefaults().
		OrderBy("Article.Title", liqu.Asc).
		Select("Article", "*").
		Select("Author", "*").
		Select("Categories", "*")

	// initiate a new instance of liqu
	li := liqu.New(context.TODO(), nil).
		WithDefaults(def)

	// pass the object to liqu and all magic happens here
	err := li.FromSource(&list)
	if err != nil {
		log.Fatal(err)
		return
	}

	// get the SQL Query and SQL Params
	sqlQuery, sqlParams := li.SQL()
	liqu.Debug(sqlQuery, sqlParams)

	var result string
	// now you can pass the sqlQuery and sqlParams to your favorite sql executor.
	// result = sql.SelectString(sqlQuery, sqlParams...)

	// the PostProcess method helps to filter out the row count
	// and removes it from the json all together
	result = li.PostProcess(result)

	// after the PostProcess you kan fetch the paging params
	listFilters := li.Filters()
	liqu.Debug(listFilters)

	// at this point, the result is a valid JSON string which
	// can be used as api output. the json only contains the selected fields.
	// alternatively, you can marshal the json back in the original struct.
	err = json.Unmarshal([]byte(result), &list)
	if err != nil {
		log.Fatal(err)
		return
	}
}
