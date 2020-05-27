package robert

import "time"

type Query struct {
	Id          string
	Query       string
	Lifetime    time.Duration
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastRefresh time.Time
}

type CreateQuery struct {
	Query    string
	Lifetime time.Duration
}

type UpdateQuery struct {
	Query    *string
	Lifetime *time.Duration
}
