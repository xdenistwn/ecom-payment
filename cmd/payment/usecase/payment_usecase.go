package usecase

import (
	"context"
	"errors"
	"fmt"
	"payment/cmd/payment/service"
	"payment/infrastructure/constant"
	"payment/infrastructure/log"
	"payment/models"
	"payment/pdf"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type PaymentUsecase interface {
	ProcessPaymentWebhook(ctx context.Context, payload models.XenditWebhookPayload) error
	ProcessPaymentRequest(ctx context.Context, payload models.OrderCreatedEvent) error
	DownloadPDFInvoice(ctx context.Context, orderID int64) (string, error)
}

type paymentUsecase struct {
	Service service.PaymentService
}

func NewPaymentUsecase(svc service.PaymentService) PaymentUsecase {
	return &paymentUsecase{
		Service: svc,
	}
}

func (uc *paymentUsecase) ProcessPaymentRequest(ctx context.Context, payload models.OrderCreatedEvent) error {
	err := uc.Service.SavePaymentRequest(ctx, models.PaymentRequests{
		OrderID:    payload.OrderID,
		Amount:     payload.TotalAmount,
		UserID:     payload.UserID,
		Status:     "PENDING",
		CreateTime: time.Now(),
	})
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"payload": payload,
		}).Errorf("uc.svc.SavePaymentRequest() got error: %v", err)

		return err
	}

	return nil
}

func (uc *paymentUsecase) DownloadPDFInvoice(ctx context.Context, orderID int64) (string, error) {
	paymentDetail, err := uc.Service.GetPaymentInfoByOrderID(ctx, orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("uc.svc.GetPaymentInfoByOrderID() got error: %v", err)

		return "", err
	}

	filePath := fmt.Sprintf("/invoices/invoice_%d.pdf", orderID)

	err = pdf.GenerateInvoicePDF(*paymentDetail, filePath)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
			"file":     filePath,
		}).Errorf("pdf.GenerateInvoicePDF() got error: %v", err)

		return "", err
	}

	return filePath, nil
}

func (uc *paymentUsecase) ProcessPaymentWebhook(ctx context.Context, payload models.XenditWebhookPayload) error {
	switch payload.Status {
	case "PAID":
		orderID := extractExternalIDToOrderId(payload.ExternalID)

		// validate webhook amount before process payment success
		amount, err := uc.Service.CheckPaymentAmountByOrderID(ctx, orderID)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"order_id":       orderID,
				"status":         payload.Status,
				"external_id":    payload.ExternalID,
				"webhook_amount": payload.Amount,
			}).Errorf("uc.svc.ProcessPaymentSuccess() got error: %v", err)

			return err
		}

		if amount != payload.Amount {
			// insert into payment anomaly table
			errorInvalidAmount := fmt.Sprintf("Webhook amount mismatch: expected %.2f, got %.2f", amount, payload.Amount)
			paymentAnomaly := models.PaymentAnomaly{
				OrderID:     orderID,
				ExternalID:  payload.ExternalID,
				AnomalyType: constant.AnomalyTypeInvalidAmount,
				Notes:       errorInvalidAmount,
				Status:      constant.PaymentAnomalyStatusNeedToCheck,
				CreateTime:  time.Now(),
			}

			err = uc.Service.SavePaymentAnomaly(ctx, paymentAnomaly)
			if err != nil {
				log.Logger.WithFields(logrus.Fields{
					"payload":        payload,
					"paymentAnomaly": paymentAnomaly,
				}).WithError(err)

				return err
			}

			log.Logger.WithFields(logrus.Fields{
				"payload": payload,
			}).Error(errorInvalidAmount)
			err = errors.New(errorInvalidAmount)

			return err
		}

		err = uc.Service.ProcessPaymentSuccess(ctx, orderID)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"status":      payload.Status,
				"external_id": payload.ExternalID,
			}).Errorf("uc.svc.ProcessPaymentSuccess() got error: %v", err)

			return err
		}
	case "FAILED":
	case "PENDING":
	default:
		log.Logger.WithFields(logrus.Fields{
			"status":      payload.Status,
			"external_id": payload.ExternalID,
		}).Infof("[%s] Anomaly Payment Webhook Status not found: %s", payload.ExternalID, payload.Status)

		// maybe store to payment_anomaly table, so we can proceed manually later.
	}

	return nil
}

func extractExternalIDToOrderId(externalID string) int64 {
	// sample
	// key kafka event: "order-12345"
	idStr := strings.TrimPrefix(externalID, "order-")
	orderId, _ := strconv.ParseInt(idStr, 10, 64)

	return orderId
}
