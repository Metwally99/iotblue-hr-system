package model

import "time"

type Department struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
