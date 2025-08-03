package models

import "time"

type PaymentAnomaly struct {
	ID          int64     `json:"id"`
	OrderID     int64     `json:"order_id"`
	ExternalID  string    `json:"external_id"`
	AnomalyType int       `json:"anomaly_type"`
	Notes       string    `json:"notes"`
	Status      int       `json:"status"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}
