# liqu

- is short for *List Query*
- transforms nested structs into sql queries
- provides search & pagination
- is build to work with `postgres` and relies heavy on `json` & `outer join`

## Interface

all structs that represent a table should inherit the Source interface

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
	"github.com/donseba/liqu"
	"log"
	"time"
)

type (
	Article struct {
		ID         int       `db:"id"`
		Title      string    `db:"title"`
		Content    string    `db:"content"`
		Date       time.Time `db:"date"`
		AuthorID   int       `db:"author_id"`
		CategoryID int       `db:"category_id"`
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

	// after the PostProcess you kan fetch the paging params
	filters := li.Filters()
	liqu.Debug(filters)
}
```

which results in the following base query:

```postgresql
 SELECT    SELECT coalesce(jsonb_agg(q),'[]') FROM ( 
    SELECT Count(*) over()     AS totalrows,
           To_jsonb( article ) AS article,
           author.author       AS "Author",
           category.category   AS "Category"
    FROM       (
        SELECT article.id AS "ID"
        FROM   article ) AS article
        right join lateral (
              SELECT to_jsonb( jsonb_build_object( 'ID', author.id ) ) AS author
              FROM   author
              WHERE  id = article."AuthorID" 
        ) AS author ON TRUE
        left join  lateral (
              SELECT to_jsonb( jsonb_build_object( 'ID', author.id ) ) AS category
              FROM   author
              WHERE  id = article."CategoryID" 
        ) AS category ON TRUE 
    limit 25 offset 0 
) q

```

## TODO

- [ ]  CTE
- [x]  Order by
- [x]  Specify fields to select
- [ ]  Add option to default to all fields
- [ ]  Aggregate functions like SUM, AVG, MIN, MAX
- [ ]  sub query support into a single field
- [x]  default order by
- [x]  protected where clause, to force company uuid or other value
- [ ]  tests
