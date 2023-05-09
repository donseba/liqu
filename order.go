package liqu

import (
	"fmt"
	"strings"
)

type OrderDirection string

const (
	Asc  OrderDirection = "ASC"
	Desc                = "DESC"
)

func (od OrderDirection) String() string {
	return string(od)
}

type OrderBuilder struct {
	orders []string
}

func NewOrderBuilder() *OrderBuilder {
	return &OrderBuilder{
		orders: []string{},
	}
}

func (ob *OrderBuilder) OrderBy(column string, direction OrderDirection) *OrderBuilder {
	order := fmt.Sprintf("%s %s", column, direction)
	ob.orders = append(ob.orders, order)
	return ob
}

func (ob *OrderBuilder) Build() string {
	return strings.Join(ob.orders, ", ")
}

func ParseURLOrderQueryToOrderBuilder(query string) (*OrderBuilder, error) {
	ob := NewOrderBuilder()
	orders := strings.Split(query, ",")

	for _, order := range orders {
		parts := strings.Split(order, "|")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid order format: %s", order)
		}

		column := parts[0]
		direction := OrderDirection(parts[1])
		if direction != Asc && direction != Desc {
			return nil, fmt.Errorf("invalid order direction: %s", direction)
		}

		ob.OrderBy(column, direction)
	}

	return ob, nil
}

func (l *Liqu) parseOrderBy(query string) error {
	if strings.TrimSpace(query) == "" {
		return nil
	}

	orders := strings.Split(query, ",")

	for _, order := range orders {
		parts := strings.Split(order, "|")
		if len(parts) != 2 {
			return fmt.Errorf("invalid order format: %s", order)
		}

		var (
			model  string
			field  string
			column string
		)

		if strings.Contains(parts[0], ".") {
			el := strings.Split(parts[0], ".")
			model = el[0]
			field = el[1]
		} else {
			model = l.tree.as
			field = parts[0]
		}

		var ok bool
		if column, ok = l.registry[model].fieldDatabase[field]; !ok {
			return fmt.Errorf("invalid order field %s", parts[0])
		}

		column = fmt.Sprintf("%s.%s", l.registry[model].tableName, column)

		direction := OrderDirection(parts[1])
		if direction != Asc && direction != Desc {
			return fmt.Errorf("invalid order direction: %s", direction)
		}

		l.registry[model].branch.order.OrderBy(column, direction)
		l.registry[model].branch.groupBy.GroupBy(column)

	}

	return nil
}
