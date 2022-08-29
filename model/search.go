package model

import "strings"

const (
	ASC Sort = iota
	DES
)

const (
	DATE By = iota
	PRICE
	EST
	STATUS
	TYPE
)

type Sort uint
type By uint

type OrderBy struct {
	By   By   `json:"by,omitempty"`
	Sort Sort `json:"sort,omitempty"`
}

type Search struct {
	OrderBys []OrderBy `json:"order,omitempty"`

	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type SearchOrder struct {
	Search
	Status []Status `json:"status"`
	Ests   []uint64 `json:"ests,omitempty"`
	Users  []uint64
	Types  []Type
	Lower  float64
	Higher float64
}

func (o OrderBy) get() string {
	var sort string
	var order string
	var b strings.Builder

	switch o.By {
	case DATE:
		order = "created_at"
	case TYPE:
		order = "type_id"
	case PRICE:
		order = "total"
	case EST:
		order = "establishment_id"
	case STATUS:
		order = "status_id"
	default:
		return ""
	}

	if o.Sort == DES {
		sort = " DESC"
	}
	b.WriteString(order)
	b.WriteString(sort)
	return b.String()
}

func (s Search) Query() string {
	var q strings.Builder
	for i, o := range s.OrderBys {
		g := o.get()
		if i != 0 && g != "" {
			q.WriteString(",")
		}
		q.WriteString(g)
	}
	return q.String()
}
