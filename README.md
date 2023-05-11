# liqu

- is short for *List Query*
- transforms nested structs into sql queries
- provides search & pagination out of the box
- is build to work with `postgres` and relies heavily on `jsonb` & `lateral joins`
- outputs into json and can marshal back into the initial source

## Note 
Liqu is in an early phase, and the API might change over time.
However, the goal is to build the initial main release around the current API and push out a first stable release.

## Interface

all structs that represent a table need to inherit the Source interface

```go
Source interface {  
    Table() string  
    PrimaryKeys() []string  
}
```

like so:

```go
type Article struct {  
    ID       int    `db:"id"`  
    Title    string `db:"title"`  
    Content  string `db:"content"`  
    AuthorID int    `db:"author_id"`  
}

func (*Article) Table() string {  
    return "article"
}  
  
func (*Article) PrimaryKeys() []string {  
    return []string{"ID"}
}
```

## Example

```go
package main

import (
	"context"
	"encoding/json"
	"github.com/donseba/liqu"
	"log"
	"time"
)

type (
	Article struct {
		ID            int       `db:"id"`
		Title         string    `db:"title"`
		Content       string    `db:"content"`
		Date          time.Time `db:"date"`
		AuthorID      int       `db:"author_id"`
		CategoryID    int       `db:"category_id"`
		DocumentCount int       `db:"document_count"`
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

	// CategoryList is a list of categories with the last 5 articles
	CategoryList struct {
		Category Category
		Articles []Article `join:"left" related:"Articles.CategoryID=Category.ID" limit:"5" offset:"0" order:"Date|DESC"`
	}
)

func main() {
	// initiate a slice of the object to process
	list := make([]ArticleList, 0)

	// get the filters from the URL Query 
	filters, err := liqu.ParseUrlValuesToFilters(r.URL.Query())
	if err != nil {
		log.Fatal(err)
		return
	}

	// define defaults to be used for the query
	def := liqu.NewDefaults().
		Where("Article.ClusterID", liqu.Equal, "secret-cluster-id").
		OrderBy("Article.Title", liqu.Asc).
		Select("Article", "*").
		Select("Author", "*").
		Select("Category", "*")
	
	// it is also possible to add a subQuery and add the result into a field. 
	documentSQ := liqu.NewSubQuery("Article", "DocumentCount").
		Relate("acrticle_id", "ID").
		Select("COALESCE(SUM(1),0)").
		From("article_documents")
	
	// initiate a new instance of liqu 
	li := liqu.New(context.TODO(), filters).
		WithDefaults(def).
		WithSubQuery(documentSQ)

	// pass the object to liqu and all magic happens here
	err = li.FromSource(&list)
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
	// and removes it from the json all  together
	result = li.PostProcess(result)

	// after the PostProcess you kan fetch the paging params
	listFilters := li.Filters()
	liqu.Debug(listFilters)

	// at this point the result is a valid JSON string which 
	// can be used as api output. the json only contains the selected fields. 
	// alternatively you can marshal the json back in the original struct.
	err = json.Unmarshal([]byte(result), &list)
	if err != nil {
		log.Fatal(err)
		return
	}
}
```

which results in the following base query (without the defaults):

```postgresql
SELECT
    coalesce(jsonb_agg(q),'[]')
FROM (
    SELECT
        count(*) OVER() AS TotalRows,
        to_jsonb("Article") AS "Article",
        "Author"."Author" AS "Author",
        "Category"."Category" AS "Category"
    FROM  (
        SELECT "article"."id" AS "ID"
        FROM "article"
        GROUP BY "article"."id"
    ) AS "Article"
    RIGHT JOIN LATERAL (
        SELECT to_jsonb( jsonb_build_object('ID', "author"."id") ) AS "Author"
        FROM "author"
        WHERE id = "Article"."AuthorID"
    ) AS "Author" ON true
    LEFT JOIN LATERAL (
        SELECT to_jsonb( jsonb_build_object('ID', "author"."id") ) AS "Category"
        FROM "author"
        WHERE id = "Article"."CategoryID"
    ) AS "Category" ON true
    LIMIT 25 OFFSET 0
) q

```

## TODO

- [ ]  CTE
- [ ]  Aggregate functions like SUM, AVG, MIN, MAX
- [ ]  tests
- [x]  sub query support into a single field
- [x]  Order by
- [x]  Specify fields to select
- [x]  Add option to default to all fields
- [x]  default order by
- [x]  protected where clause, to force company uuid or other value

## About 
liqu is born from the need to be able to quickly create paginated results with filtering capabilities.

I've created a similar package while working for a previous employer,
and during my 8 years we made 3 iterations which all where pretty similar to each other.
Building this one from scratch with new insights and knowledge was a fun project.
According to by knowledge, there is no tool available in the golang ecosystem which can do this.

Some coding styles and usages are inspired by Doug Martin's goqu package. 

There are obviously some limitations to what this can and cannot do, mainly because I didn't have a use case for it. 