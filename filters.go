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

func (f *Filters) First() string {
	uv := url.Values{}

	uv.Set("page", fmt.Sprintf("%v", 1))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) Previous() string {
	uv := url.Values{}

	if f.Page <= 1 {
		return ""
	}

	uv.Set("page", fmt.Sprintf("%v", 1))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) Current() string {
	uv := url.Values{}

	uv.Set("page", fmt.Sprintf("%v", f.Page))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) Next() string {
	uv := url.Values{}

	if f.Page >= f.totalPages {
		return ""
	}

	uv.Set("page", fmt.Sprintf("%v", f.Page+1))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) Last() string {
	uv := url.Values{}

	if f.Page == f.totalPages {
		return ""
	}

	uv.Set("page", fmt.Sprintf("%v", f.totalPages))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) params(query url.Values) string {
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
