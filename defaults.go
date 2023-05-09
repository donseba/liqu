package liqu

type (
	Defaults struct {
		where   map[string]defaultWhere
		orderBy map[string]OrderDirection
		sel     map[string][]string
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
		sel:     make(map[string][]string),
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

func (d *Defaults) Select(model string, fields ...string) *Defaults {
	if d.sel[model] == nil {
		d.sel[model] = make([]string, 0)
	}

	d.sel[model] = append(d.sel[model], fields...)

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

	for model, fields := range l.defaults.sel {
		for _, field := range fields {
			l.processSelect(model, field)
		}
	}

	return nil
}
