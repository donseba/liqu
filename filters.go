package liqu

import (
	"fmt"
	"html/template"
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

	Range struct {
		Number  int
		Current bool
		Url     template.URL
	}

	Ranges []Range
)

func (f *Filters) TotalResults() int {
	return f.totalResults
}

func (f *Filters) TotalPages() int {
	return f.totalPages
}

func (f *Filters) FirstOnPage() int {
	if f.Page == 1 {
		return 1
	}

	return ((f.Page - 1) * f.PerPage) + 1
}

func (f *Filters) LastOnPage() int {
	var (
		total = f.totalResults
	)

	calc := f.Page * f.PerPage

	if calc < total {
		total = calc
	}

	return total
}

func (f *Filters) First() template.URL {
	uv := url.Values{}

	uv.Set("page", fmt.Sprintf("%v", 1))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) Previous() template.URL {
	uv := url.Values{}

	if f.Page <= 1 {
		return ""
	}

	uv.Set("page", fmt.Sprintf("%v", f.Page-1))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) Current() template.URL {
	uv := url.Values{}

	uv.Set("page", fmt.Sprintf("%v", f.Page))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) Next() template.URL {
	uv := url.Values{}

	if f.Page >= f.totalPages {
		return ""
	}

	uv.Set("page", fmt.Sprintf("%v", f.Page+1))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) Last() template.URL {
	uv := url.Values{}

	if f.Page == f.totalPages {
		return ""
	}

	uv.Set("page", fmt.Sprintf("%v", f.totalPages))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) Range() Ranges {
	r := make(Ranges, 0)

	for i := f.Page - 5; i < f.Page+5; i++ {
		if i <= 0 || i > f.totalPages {
			continue
		}

		r = append(r, Range{
			Number:  i,
			Current: i == f.Page,
			Url:     f.set(i),
		})
	}

	return r
}

func (f *Filters) set(page int) template.URL {
	uv := url.Values{}

	uv.Set("page", fmt.Sprintf("%v", page))
	uv.Set("per_page", fmt.Sprintf("%v", f.PerPage))

	return f.params(uv)
}

func (f *Filters) params(query url.Values) template.URL {
	if strings.TrimSpace(f.Where) != "" {
		query.Set("where", f.Where)
	}

	if strings.TrimSpace(f.OrderBy) != "" {
		query.Set("order_by", f.OrderBy)
	}

	if strings.TrimSpace(f.Select) != "" {
		query.Set("select", f.Select)
	}

	return template.URL(query.Encode())
}
