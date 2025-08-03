package usecase

import (
	"context"
	"payment/cmd/payment/service"
	"payment/infrastructure/log"
	"payment/models"

	"github.com/sirupsen/logrus"
)

type XenditUsecase interface {
	CreateInvoice(ctx context.Context, param models.OrderCreatedEvent) error
}

type xenditUsecase struct {
	xenditService service.XenditService
}

func NewXenditUsecase(xenditService service.XenditService) XenditUsecase {
	return &xenditUsecase{
		xenditService: xenditService,
	}
}

func (uc *xenditUsecase) CreateInvoice(ctx context.Context, param models.OrderCreatedEvent) error {
	err := uc.xenditService.CreateInvoice(ctx, param)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("CreateInvoice => uc.xenditService.CreateInvoice got error: %v", err)

		return err
	}

	return nil
}
