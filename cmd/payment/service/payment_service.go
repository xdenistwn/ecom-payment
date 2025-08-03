package service

import (
	"context"
	"math"
	"payment/cmd/payment/repository"
	"payment/infrastructure/constant"
	"payment/infrastructure/log"
	"payment/models"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	MaxTryPublishPayment = 3
	RetryDelay           = 2 // seconds
)

type PaymentService interface {
	ProcessPaymentSuccess(ctx context.Context, orderID int64) error
	CheckPaymentAmountByOrderID(ctx context.Context, orderID int64) (float64, error)
	SavePaymentAnomaly(ctx context.Context, param models.PaymentAnomaly) error
	SavePaymentRequest(ctx context.Context, param models.PaymentRequests) error
	GetPaymentInfoByOrderID(ctx context.Context, orderID int64) (*models.Payment, error)
}

type paymentService struct {
	database  repository.PaymentDatabase
	publisher repository.PaymentEventPublisher
}

func NewPaymentService(db repository.PaymentDatabase, publisher repository.PaymentEventPublisher) PaymentService {
	return &paymentService{
		database:  db,
		publisher: publisher,
	}
}

func (s *paymentService) CheckPaymentAmountByOrderID(ctx context.Context, orderID int64) (float64, error) {
	amount, err := s.database.CheckPaymentAmountByOrderID(ctx, orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("s.database.CheckPaymentAmountByOrderID() got error: %v", err)

		return 0, err
	}

	return amount, nil
}

func (s *paymentService) SavePaymentAnomaly(ctx context.Context, param models.PaymentAnomaly) error {
	err := s.database.SavePaymentAnomaly(ctx, param)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("s.database.SavePaymentAnomaly() got error: %v", err)

		return err
	}

	return nil
}

func (s *paymentService) SavePaymentRequest(ctx context.Context, param models.PaymentRequests) error {
	err := s.database.SavePaymentRequest(ctx, param)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("s.database.SavePaymentRequest() got error: %v", err)

		return err
	}

	return nil
}

func (s *paymentService) GetPaymentInfoByOrderID(ctx context.Context, orderID int64) (*models.Payment, error) {
	paymentInfo, err := s.database.GetPaymentInfoByOrderID(ctx, orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("s.database.GetPaymentInfoByOrderID() got error: %v", err)

		return nil, err
	}

	return paymentInfo, nil
}

func (s *paymentService) ProcessPaymentSuccess(ctx context.Context, orderID int64) error {
	// validate paid status
	isAlreadyPaid, err := s.database.IsAlreadyPaid(ctx, orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("s.database.IsAlreadyPaid() got error: %v", err)

		return err
	}

	if isAlreadyPaid {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Infof("Payment %d already paid.", orderID)

		return nil
	}

	// public event to kafka
	err = retryPublishPayment(MaxTryPublishPayment, func() error {
		errLogAudit := s.database.InsertAuditLog(ctx, models.PaymentAuditLog{
			OrderID:    orderID,
			Event:      "PublishPaymentSuccess",
			Actor:      "payment_service",
			CreateTime: time.Now(),
		})
		if errLogAudit != nil {
			log.Logger.WithFields(logrus.Fields{
				"order_id": orderID,
			}).Errorf("s.database.InsertAuditLog() got error: %v", errLogAudit)
		}

		return s.publisher.PublishPaymentSuccess(ctx, orderID)
	})
	if err != nil {
		// store data to failed payment event
		failedEventParam := models.FailedEvents{
			OrderID:    orderID,
			FailedType: constant.FailedPublishEventPaymentSuccess,
			Status:     constant.FailedPublishEventStatusNeedToCheck,
			Notes:      err.Error(),
			CreateTime: time.Now(),
		}

		// its also called dead letter queue
		errSaveFailedPublish := s.database.SaveFailedPublishEvent(ctx, failedEventParam)
		if errSaveFailedPublish != nil {
			log.Logger.WithFields(logrus.Fields{
				"failedEventParam": failedEventParam,
			}).WithError(errSaveFailedPublish).Error("s.database.SaveFailedPublishEvent() got error")

			return errSaveFailedPublish
		}

		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("s.publisher.PublishPaymentSuccess() got error: %v", err)

		return err
	}

	// update status to DB
	err = s.database.MarkPaid(ctx, orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("s.database.MarkPaid got error: %v", err)

		return err
	} else {
		errLogAudit := s.database.InsertAuditLog(ctx, models.PaymentAuditLog{
			OrderID:    orderID,
			Event:      "MarkPaid",
			Actor:      "payment_service",
			CreateTime: time.Now(),
		})
		if errLogAudit != nil {
			log.Logger.WithFields(logrus.Fields{
				"order_id": orderID,
			}).WithError(errLogAudit).Errorf("s.database.InsertAuditLog() got error: %v", errLogAudit)
		}
	}

	return nil
}

func retryPublishPayment(max int, fn func() error) error {
	var err error
	for i := 0; i < max; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		wait := time.Duration(math.Pow(2, float64(i))) * time.Second
		log.Logger.Printf("Retrying to publish payment, attempt %d/%d, waiting %s: %v", i+1, max, wait, err)
		time.Sleep(wait)
	}

	return err
}
