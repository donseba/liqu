package liqu

import (
	"testing"
)

func TestWhereClause(t *testing.T) {
	where := &WhereClause{}
	where.AddCondition("name", ILike, "John%")
	where.AddCondition("email", Like, "john@%")
	where.AddCondition("age", GreaterThanOrEqual, 25)
	where.AddCondition("age", LessThan, 65)
	where.AddCondition("city", IsNotNull, nil)
	where.AddCondition("score", Between, []interface{}{50, 100})
	where.AddCondition("country", In, []interface{}{"USA", "UK", "Canada"})
	where.AddCondition("job", NotLike, "manager%")
	where.AddCondition("title", NotILike, "director%")
	where.AddCondition("salary", NotEqual, 100000)
	where.AddCondition("salary", NotEqualAlt, 100000)
	where.AddCondition("start_date", LessThanOrEqual, "2022-01-01")
	where.AddCondition("end_date", GreaterThan, "2023-01-01")
	where.AddCondition("nickname", StartsWith, "A")
	where.AddCondition("notes", Any, "'{John, Jane, Jack}'::text[]")
	where.AddCondition("position", NotAny, "'{Manager, Director}'::text[]")

	nestedOr := []*Condition{
		{Field: "department", Operator: Equal, Value: "HR"},
		{Field: "department", Operator: Equal, Value: "IT"},
	}
	where.AddNestedCondition(Or, nestedOr...)

	nestedNotIn := []*Condition{
		{Field: "role", Operator: NotIn, Value: []interface{}{"admin", "superadmin"}},
	}
	where.AddNestedCondition(And, nestedNotIn...)

	expected := `WHERE name ILIKE 'John%' AND email LIKE 'john@%' AND age >= 25 AND age < 65 AND city IS NOT NULL  AND score BETWEEN (50, 100) AND country IN ('USA', 'UK', 'Canada') AND job NOT LIKE 'manager%' AND title NOT ILIKE 'director%' AND salary <> 100000 AND salary != 100000 AND start_date <= '2022-01-01' AND end_date > '2023-01-01' AND nickname ^ 'A' AND notes ANY ''{John, Jane, Jack}'::text[]' AND position NOT ANY ''{Manager, Director}'::text[]' OR (department = 'HR' OR department = 'IT') AND (role NOT IN ('admin', 'superadmin'))`
	result := where.Build()

	if expected != where.Build() {
		t.Errorf("expected:\n%s\ngot:\n %s", expected, result)
	}
}
