package liqu

type (
	Defaults struct {
		where   map[string]defaultWhere
		orderBy map[string]OrderDirection
		sel     []string
	}

	defaultWhere struct {
		column string
		op     Operator
		val    interface{}
	}
)

func NewDefaults() *Defaults {
	return &Defaults{
		where:   make(map[string]defaultWhere),
		orderBy: make(map[string]OrderDirection),
		sel:     make([]string, 0),
	}
}

func (d *Defaults) OrderBy(column string, direction OrderDirection) *Defaults {
	d.orderBy[column] = direction

	return d
}

func (d *Defaults) Where(column string, op Operator, value interface{}) *Defaults {
	d.where[column] = defaultWhere{
		column: column,
		op:     op,
		val:    value,
	}

	return d
}
func (d *Defaults) Select(column string) *Defaults {
	for _, v := range d.sel {
		if v == column {
			return d
		}
	}

	d.sel = append(d.sel, column)

	return d
}

func (l *Liqu) processDefaults() error {
	for k, v := range l.defaults.orderBy {
		err := l.processOrderBy(k, v.String())
		if err != nil {
			return err
		}
	}

	for _, v := range l.defaults.where {
		err := l.processWhere(And, v.column, v.op.String(), v.val, true)
		if err != nil {
			return err
		}
	}

	return nil
}
