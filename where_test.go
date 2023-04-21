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

	expected := `name ILIKE '%John%' AND age >= 18 OR age IS NULL AND (country = 'USA' OR city = 'New York')`

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
			n.Column("country").In("USA", "Canada").
				OrNotIn("city", "New York", "Los Angeles")
		}).
		Build()

	expected := `name ILIKE '%John%' AND age >= 18 OR age IS NULL AND (USA IN ('Canada') OR city NOT IN ('New York', 'Los Angeles'))`
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
	fmt.Println("WHERE", whereClause)
}
