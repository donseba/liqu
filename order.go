package liqu

import (
	"fmt"
	"strings"
)

const (
	Asc  OrderDirection = "ASC"
	Desc                = "DESC"
)

type (
	OrderDirection string

	OrderBuilder struct {
		orders []Order
	}

	Order struct {
		Column    string
		Direction OrderDirection
	}
)

func (od OrderDirection) String() string {
	return string(od)
}

func NewOrderBuilder() *OrderBuilder {
	return &OrderBuilder{
		orders: []Order{},
	}
}

func (ob *OrderBuilder) OrderBy(column string, direction OrderDirection) *OrderBuilder {
	ob.orders = append(ob.orders, Order{
		Column:    column,
		Direction: direction,
	})

	return ob
}

func (ob *OrderBuilder) HasOrderBy(column string) bool {
	for _, v := range ob.orders {
		if v.Column == column {
			return true
		}
	}

	return false
}
func (ob *OrderBuilder) Unset(column string) {
	for k, v := range ob.orders {
		if v.Column == column {
			ob.orders = append(ob.orders[:k], ob.orders[k+1:]...)
			return
		}
	}

	return
}

func (ob *OrderBuilder) Build() string {
	parts := make([]string, 0)
	for _, v := range ob.orders {
		parts = append(parts, fmt.Sprintf("%s %s", v.Column, v.Direction))
	}

	return strings.Join(parts, ", ")
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

		err := l.processOrderBy(parts[0], parts[1])
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Liqu) processOrderBy(col, dir string) error {
	var (
		model  string
		field  string
		column string
	)

	if strings.Contains(col, ".") {
		el := strings.Split(col, ".")
		model = el[0]
		field = el[1]
	} else {
		model = l.tree.as
		field = col
	}

	var ok bool
	if column, ok = l.registry[model].fieldDatabase[field]; !ok {
		return fmt.Errorf("invalid order field %s", col)
	}

	column = fmt.Sprintf(`"%s"."%s"`, l.registry[model].tableName, column)

	if l.registry[model].branch.order.HasOrderBy(column) {
		l.registry[model].branch.order.Unset(column)
	}

	direction := OrderDirection(dir)
	if direction != Asc && direction != Desc {
		return fmt.Errorf("invalid order direction: %s", direction)
	}

	l.registry[model].branch.order.OrderBy(column, direction)
	l.registry[model].branch.groupBy.GroupBy(column)

	return nil
}
