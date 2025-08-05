package service

import (
	"context"
	"fmt"
	"payment/cmd/payment/repository"
	"payment/grpc"
	"payment/infrastructure/log"
	"payment/models"
	"time"

	"github.com/sirupsen/logrus"
)

type XenditService interface {
	CreateInvoice(ctx context.Context, param models.OrderCreatedEvent) error
}

type xenditService struct {
	database   repository.PaymentDatabase
	xendit     repository.XenditClient
	userClient grpc.UserClient
}

func NewXenditService(database repository.PaymentDatabase, xenditClient repository.XenditClient, userClient grpc.UserClient) XenditService {
	return &xenditService{
		database:   database,
		xendit:     xenditClient,
		userClient: userClient,
	}
}

func (s *xenditService) CreateInvoice(ctx context.Context, param models.OrderCreatedEvent) error {
	// get user info from user grpc service
	userInfo, err := s.userClient.GetUserInfoByUserId(ctx, param.UserID)
	if err != nil {
		return err
	}

	externalID := fmt.Sprintf("order-%d", param.OrderID)
	req := models.XenditInvoiceRequest{
		ExternalID:  externalID,
		Amount:      param.TotalAmount,
		Description: fmt.Sprintf("Pembayaran Order %d", param.OrderID),
		PayerEmail:  userInfo.Email,
	}

	xenditInvoice, err := s.xendit.CreateInvoice(ctx, req)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param":   param,
			"payload": req,
		}).Errorf("s.xendit.CreateInvoice() got error: %v", err)

		return err
	}

	// save to DB
	newPayment := models.Payment{
		OrderID:     param.OrderID,
		UserID:      param.UserID,
		ExternalID:  externalID,
		Amount:      param.TotalAmount,
		Status:      "PENDING",
		ExpiredTime: xenditInvoice.ExpiryDate,
		CreateTime:  time.Now(),
	}
	err = s.database.SavePayment(ctx, newPayment)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param":      param,
			"newPayment": newPayment,
		}).Errorf("CreateInvoice => s.database.SavePayment got error: %v", err)

		return err
	}

	return nil
}
