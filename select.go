package liqu

import (
	"fmt"
	"strings"
)

func (l *Liqu) parseSelect(query string) error {
	if strings.TrimSpace(query) == "" {
		return nil
	}

	selects := strings.Split(query, ",")

	for _, sel := range selects {
		parts := strings.Split(sel, ".")
		if len(parts) != 2 {
			return fmt.Errorf("invalid order format: %s", sel)
		}

		var (
			model = parts[0]
			field = parts[1]
		)

		if field == "*" {
			for field, _ := range l.registry[model].fieldTypes {
				l.registry[model].branch.selectedFields[field] = true
			}
		} else {
			l.registry[model].branch.selectedFields[field] = true
		}
	}

	return nil
}
