package types

import "time"

type Goods struct {
	ID          int       `json:"id"`
	ProjectID   int       `json:"campaignId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"createdAt"`
}

type GoodsLog struct {
	ID          int       `json:"id"`
	ProjectID   int       `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Removed     bool      `json:"removed"`
	EventTime   time.Time `json:"eventTime"`
}

type Meta struct {
	Total   int `json:"total"`
	Removed int `json:"removed"`
	Limit   int `json:"limit"`
	Offset  int `json:"offset"`
}

type List struct {
	Meta  Meta    `json:"meta"`
	Goods []Goods `json:"goods"`
}

type Priority struct {
	ID       int `json:"id"`
	Priority int `json:"priority"`
}

type Priorities struct {
	Priorities []Priority `json:"priorities"`
}
