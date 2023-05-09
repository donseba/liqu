package liqu

import (
	"fmt"
	"net/url"
	"strings"
)

type (
	Filters struct {
		Page          int
		PerPage       int
		totalResults  int
		totalPages    int
		DisablePaging bool
		Where         string
		OrderBy       string
		Select        string
	}
)

func (f *Filters) Query() string {
	query := url.Values{}

	query.Set("page", fmt.Sprintf("%v", f.Page))
	query.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	if strings.TrimSpace(f.Where) != "" {
		query.Set("where", f.Where)
	}

	if strings.TrimSpace(f.OrderBy) != "" {
		query.Set("order_by", f.OrderBy)
	}

	if strings.TrimSpace(f.Select) != "" {
		query.Set("select", f.Select)
	}

	return query.Encode()
}
