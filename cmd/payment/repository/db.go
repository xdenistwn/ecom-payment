package repository

import (
	"context"
	"payment/infrastructure/log"
	"payment/models"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PaymentDatabase interface {
	MarkPaid(ctx context.Context, orderID int64) error
	MarkExpired(ctx context.Context, paymentID int64) error
	SavePayment(ctx context.Context, param models.Payment) error
	IsAlreadyPaid(ctx context.Context, orderID int64) (bool, error)
	CheckPaymentAmountByOrderID(ctx context.Context, orderID int64) (float64, error)
	SavePaymentAnomaly(ctx context.Context, param models.PaymentAnomaly) error
	SaveFailedPublishEvent(ctx context.Context, param models.FailedEvents) error
	SavePaymentRequest(ctx context.Context, param models.PaymentRequests) error
	GetPendingInvoices(ctx context.Context) ([]models.Payment, error)
	GetPaymentInfoByOrderID(ctx context.Context, orderID int64) (*models.Payment, error)
	GetPendingPaymentRequests(ctx context.Context, paymentRequests *[]models.PaymentRequests) error
	GetFailedPaymentRequests(ctx context.Context, paymentRequests *[]models.PaymentRequests) error
	GetExpiredPendingPayments(ctx context.Context) ([]models.Payment, error)
	UpdateSuccessPaymentRequest(ctx context.Context, paymentRequestID int64) error
	UpdatePendingPaymentRequest(ctx context.Context, paymentRequestID int64) error
	UpdateFailedPaymentRequest(ctx context.Context, paymentRequestID int64, notes string) error

	// audit logs
	InsertAuditLog(ctx context.Context, param models.PaymentAuditLog) error
}

type paymentDatabase struct {
	DB *gorm.DB
}

func NewPaymentDatabase(db *gorm.DB) PaymentDatabase {
	return &paymentDatabase{
		DB: db,
	}
}

func (r *paymentDatabase) CheckPaymentAmountByOrderID(ctx context.Context, orderID int64) (float64, error) {
	var result models.Payment
	err := r.DB.Table("payments").WithContext(ctx).Where("order_id = ?", orderID).First(&result).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("Repository => CheckPaymentAmountByOrderID got error: %v", err)

		return 0, err
	}

	return result.Amount, nil
}

func (r *paymentDatabase) MarkPaid(ctx context.Context, orderID int64) error {
	err := r.DB.Model(&models.Payment{}).Table("payments").WithContext(ctx).Where("order_id = ?", orderID).Update("status", "paid").Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("MarkPaid => r.DB.Update() MarkPaid got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) GetPaymentInfoByOrderID(ctx context.Context, orderID int64) (*models.Payment, error) {
	var payment models.Payment
	err := r.DB.Table("payments").WithContext(ctx).Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("GetPaymentInfoByOrderID => r.DB.First() got error: %v", err)

		return &models.Payment{}, err
	}

	return &payment, nil
}

func (r *paymentDatabase) SavePayment(ctx context.Context, param models.Payment) error {
	err := r.DB.Create(param).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("SavePayment => r.DB.Create() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) SaveFailedPublishEvent(ctx context.Context, param models.FailedEvents) error {
	err := r.DB.Table("failed_events").WithContext(ctx).Create(param).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("SaveFailedPublishEvent => r.DB.Create() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) IsAlreadyPaid(ctx context.Context, orderID int64) (bool, error) {
	var result models.Payment
	err := r.DB.Table("payments").WithContext(ctx).Where("external_id = ?", orderID).First(&result).Error
	if err != nil {
		return false, err
	}

	return result.Status == "PAID", nil
}

func (r *paymentDatabase) GetPendingInvoices(ctx context.Context) ([]models.Payment, error) {
	var result []models.Payment
	err := r.DB.Table("payments").WithContext(ctx).Where("status = ? AND create_time >= now() - interval '1 day'", "PENDING").Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *paymentDatabase) SavePaymentAnomaly(ctx context.Context, param models.PaymentAnomaly) error {
	err := r.DB.Table("payment_anomalies").WithContext(ctx).Create(param).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("SavePaymentAnomaly => r.DB.Create() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) SavePaymentRequest(ctx context.Context, param models.PaymentRequests) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Create(models.PaymentRequests{
		OrderID:    param.OrderID,
		UserID:     param.UserID,
		Amount:     param.Amount,
		Status:     param.Status,
		CreateTime: param.CreateTime,
	}).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("SavePaymentRequest => r.DB.Create() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) GetPendingPaymentRequests(ctx context.Context, paymentRequests *[]models.PaymentRequests) error {
	// batch size of 5
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("status = ?", "PENDING").Order("create_time ASC").Limit(5).Find(paymentRequests).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Errorf("GetPendingPaymentRequests => r.DB.Find() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) GetFailedPaymentRequests(ctx context.Context, paymentRequests *[]models.PaymentRequests) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("status = ?", "FAILED").Where("retry_count <= ?", 3).Order("create_time ASC").Limit(5).Find(paymentRequests).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Errorf("GetFailedPaymentRequests => r.DB.Find() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) UpdateSuccessPaymentRequest(ctx context.Context, paymentRequestID int64) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("id = ?", paymentRequestID).Updates(map[string]interface{}{
		"status":      "SUCCESS",
		"update_time": time.Now(),
	}).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"id":          paymentRequestID,
			"status":      "SUCCESS",
			"update_time": time.Now(),
		}).Errorf("UpdateSuccessPaymentRequest => r.DB.Update() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) UpdateFailedPaymentRequest(ctx context.Context, paymentRequestID int64, notes string) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("id = ?", paymentRequestID).Updates(map[string]interface{}{
		"status":      "FAILED",
		"notes":       notes,
		"retry_count": gorm.Expr("retry_count + 1"),
		"update_time": time.Now(),
	}).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"id":          paymentRequestID,
			"status":      "FAILED",
			"update_time": time.Now(),
			"notes":       notes,
		}).Errorf("UpdateFailedPaymentRequest => r.DB.Update() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) UpdatePendingPaymentRequest(ctx context.Context, paymentRequestID int64) error {
	err := r.DB.Table("payment_requests").WithContext(ctx).Where("id = ?", paymentRequestID).Updates(map[string]interface{}{
		"status":      "PENDING",
		"update_time": time.Now(),
	}).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"id":          paymentRequestID,
			"status":      "PENDING",
			"update_time": time.Now(),
		}).Errorf("UpdatePendingPaymentRequest => r.DB.Update() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) GetExpiredPendingPayments(ctx context.Context) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.DB.Table("payments").WithContext(ctx).Where("status = ? AND expired_time < ?", "PENDING", time.Now()).Find(&payments).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Errorf("GetExpiredPendingPayments => r.DB.Find() got error: %v", err)

		return nil, err
	}

	return payments, nil
}

func (r *paymentDatabase) MarkExpired(ctx context.Context, paymentID int64) error {
	err := r.DB.Model(&models.Payment{}).Table("payments").WithContext(ctx).Where("id = ?", paymentID).Updates(map[string]interface{}{
		"status":      "EXPIRED",
		"update_time": time.Now(),
	}).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"id": paymentID,
		}).Errorf("MarkExpired => r.DB.Update() got error: %v", err)

		return err
	}

	return nil
}

func (r *paymentDatabase) InsertAuditLog(ctx context.Context, param models.PaymentAuditLog) error {
	err := r.DB.Table("payment_audit_logs").WithContext(ctx).Create(param).Error
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("InsertAuditLog => r.DB.Create() got error: %v", err)

		return err
	}

	return nil
}
