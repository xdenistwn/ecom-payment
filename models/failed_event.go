package models

import "time"

type FailedEvents struct {
	ID         int64     `json:"id"`
	OrderID    int64     `json:"order_id"`
	ExternalID string    `json:"external_id"`
	FailedType int       `json:"failed_type"`
	Status     int       `json:"status"`
	Notes      string    `json:"notes"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}
