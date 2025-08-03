package models

import "time"

type XenditInvoiceRequest struct {
	ExternalID  string  `json:"external_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	PayerEmail  string  `json:"payer_email"`
}

type XenditInvoiceResponse struct {
	ID         string    `json:"id"`
	ExpiryDate time.Time `json:"expiry_date"`
	InvoiceURL string    `json:"invoice_url"`
	Status     string    `json:"status"`
}
