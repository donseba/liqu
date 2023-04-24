package liqu

import (
	"fmt"
	"net/url"
	"testing"
)

func TestWhereClause(t *testing.T) {
	cb := NewConditionBuilder()
	whereClause := cb.Column("name").
		Condition(ILike, "%John%").
		And("age", GreaterThanOrEqual, 18).
		OrIsNull("age").
		AndNested(func(n *ConditionBuilder) {
			n.Column("country").Condition(Equal, "USA").
				Or("city", Equal, "New York")
		}).
		Build()

	expected := `name ILIKE $1 AND age >= $2 OR age IS NULL AND (country = $3 OR city = $4)`
	if expected != whereClause {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, whereClause)
	}
}

func TestNewConditionBuilder(t *testing.T) {
	cb := NewConditionBuilder()
	whereClause := cb.Column("name").
		Condition(ILike, "%John%").
		And("age", GreaterThanOrEqual, 18).
		OrIsNull("age").
		AndNested(func(n *ConditionBuilder) {
			n.In("country", "USA", "Canada").
				OrNotIn("city", "New York", "Los Angeles")
		}).
		And("number_seven", Equal, 7).
		Build()

	expected := `name ILIKE $1 AND age >= $2 OR age IS NULL AND (country IN ($3, $4) OR city NOT IN ($5, $6)) AND number_seven = $7`
	if expected != whereClause {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, whereClause)
	}
}

func TestParseURLQueryToConditionBuilder(t *testing.T) {
	// Simulate a URL query with the "where" parameter
	urlQuery := url.Values{}
	urlQuery.Set("where", "name|ILIKE|%John%,age|>=|18,age|IS NULL,country|IN|USA--Canada,city|NOT IN|New York--Los Angeles")

	whereQueryParam := urlQuery.Get("where")

	cb, err := ParseURLQueryToConditionBuilder(whereQueryParam)
	if err != nil {
		t.Error(err)
		return
	}

	whereClause := cb.Build()
	expected := `name ILIKE $1 AND age >= $2 AND age IS NULL AND country IN ($3, $4) AND city NOT IN ($5, $6)`
	if expected != whereClause {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, whereClause)
	}
}

func TestParseURLQueryToConditionBuilderNestedOR(t *testing.T) {
	// Simulate a URL query with the "where" parameter
	urlQuery := url.Values{}
	urlQuery.Set("where", "name|ILIKE|%John%,(OR,age|>=|18,age|IS NULL),country|IN|USA--Canada,city|NOT IN|New York--Los Angeles")

	whereQueryParam := urlQuery.Get("where")

	cb, err := ParseURLQueryToConditionBuilder(whereQueryParam)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	whereClause := cb.Build()
	expected := `name ILIKE $1 AND (age >= $2 OR age IS NULL) AND country IN ($3, $4) AND city NOT IN ($5, $6)`
	if expected != whereClause {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, whereClause)
	}

	whereArgs := cb.Args()
	expectedArgs := []interface{}{"%John%", "18", "USA", "Canada", "New York", "Los Angeles"}

	if len(whereArgs) != len(expectedArgs) {
		t.Errorf("expected:\n%d\ngot:\n%d", len(expectedArgs), len(whereArgs))
		return
	}

	for k, v := range expectedArgs {
		if v != whereArgs[k] {
			t.Errorf("expected:\n%s\ngot:\n%s", v, whereArgs[k])
		}
	}
}
