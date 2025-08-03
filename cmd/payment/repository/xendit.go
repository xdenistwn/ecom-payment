package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"payment/models"
)

type XenditClient interface {
	CreateInvoice(ctx context.Context, param models.XenditInvoiceRequest) (models.XenditInvoiceResponse, error)
	CheckInvoiceStatus(ctx context.Context, externalID string) (string, error)
}

type xenditClient struct {
	APISecretKey string
}

func NewXenditClient(apiSecretKey string) XenditClient {
	return &xenditClient{
		APISecretKey: apiSecretKey,
	}
}

func (xc *xenditClient) CreateInvoice(ctx context.Context, param models.XenditInvoiceRequest) (models.XenditInvoiceResponse, error) {
	var result models.XenditInvoiceResponse

	payload, err := json.Marshal(param)
	if err != nil {
		return models.XenditInvoiceResponse{}, err
	}

	uri := "https://api.xendit.co/v2/invoices"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewBuffer(payload))
	if err != nil {
		return models.XenditInvoiceResponse{}, err
	}

	req.SetBasicAuth(xc.APISecretKey, "")
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.XenditInvoiceResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return models.XenditInvoiceResponse{}, errors.New(fmt.Sprintf("xendit.CreateInvoice() got error %s", body))
	}

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return models.XenditInvoiceResponse{}, err
	}

	return result, nil
}

func (xc *xenditClient) CheckInvoiceStatus(ctx context.Context, externalID string) (string, error) {
	uri := fmt.Sprintf("https://api.xendit.co/v2/invoices?external_id=%s", externalID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(xc.APISecretKey, "")
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var response []models.XenditInvoiceResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response[0].Status, nil
}
