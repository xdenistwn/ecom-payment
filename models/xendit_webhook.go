package models

type XenditWebhookPayload struct {
	ExternalID string  `json:"external_id"`
	Status     string  `json:"status"`
	Amount     float64 `json:"amount"`
}
