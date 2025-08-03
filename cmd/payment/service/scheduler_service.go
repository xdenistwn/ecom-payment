package service

import (
	"context"
	"fmt"
	"log"
	"payment/cmd/payment/repository"
	"payment/models"
	"time"
)

type SchedulerService struct {
	Database       repository.PaymentDatabase
	Xendit         repository.XenditClient
	Publisher      repository.PaymentEventPublisher
	PaymentService PaymentService
}

func (s *SchedulerService) StartProcessExpiredPendingPayments() {
	ctx := context.Background()
	go func(ctx context.Context) {
		for {
			log.Println("Starting to process expired pending payments...")

			// get expired pending payments
			expiredPayments, err := s.Database.GetExpiredPendingPayments(ctx)
			if err != nil {
				log.Printf("s.Database.GetExpiredPendingPayments() got error: %v", err)
				time.Sleep(10 * time.Second) // give time gap before next iteration
				continue
			}

			for _, expiredPayment := range expiredPayments {
				err = s.Database.MarkExpired(ctx, expiredPayment.OrderID)
				if err != nil {
					log.Printf("[payment ID: %d] s.Database.MarkExpired() got error: %v", expiredPayment.ID, err)
					continue
				}
			}

			time.Sleep(10 * time.Minute) // give time gap before next iteration
		}
	}(ctx)
}

func (s *SchedulerService) StartProcessPendingPaymentRequests() {
	ctx := context.Background()
	go func(ctx context.Context) {
		for {
			var paymentRequests []models.PaymentRequests
			// get pending payment requests
			err := s.Database.GetPendingPaymentRequests(ctx, &paymentRequests)
			if err != nil {
				log.Printf("s.Database.GetPendingPaymentRequests() got error: %v", err)
				// give time gap to avoid tight loop
				time.Sleep(10 * time.Second)
				continue
			}

			for _, paymentRequest := range paymentRequests {
				// process each payment request
				externalID := fmt.Sprintf("order-%d", paymentRequest.OrderID)
				log.Printf("[DEBUG] Processing payment request ID: %d", paymentRequest.ID)

				// check if invoice has been created
				paymentInfo, err := s.Database.GetPaymentInfoByOrderID(ctx, paymentRequest.OrderID)
				if err != nil {
					log.Printf("[req id: %d] got error: %v", paymentRequest.ID, err)
					continue
				}

				if paymentInfo.ID != 0 {
					xenditInvoiceReq := models.XenditInvoiceRequest{
						ExternalID:  externalID,
						Amount:      paymentRequest.Amount,
						Description: fmt.Sprintf("Payment for Order ID %d", paymentRequest.OrderID),
						PayerEmail:  fmt.Sprintf("user%d@test.com", paymentRequest.UserID), // to do use real email
					}

					xenditInvoiceRes, err := s.Xendit.CreateInvoice(ctx, xenditInvoiceReq)
					errLogAudit := s.Database.InsertAuditLog(ctx, models.PaymentAuditLog{
						OrderID:    paymentInfo.OrderID,
						UserID:     paymentInfo.UserID,
						PaymentID:  paymentInfo.ID,
						ExternalID: paymentInfo.ExternalID,
						Event:      "CreateInvoice",
						Actor:      "scheduler_service_process_pending_payment_requests",
						CreateTime: time.Now(),
					})
					if errLogAudit != nil {
						log.Printf("[req id: %d] s.Database.InsertAuditLog() got error: %v", paymentRequest.ID, errLogAudit)
					}

					if err != nil {
						log.Printf("[req id: %d] s.Xendit.CreateInvoice() got error: %v", paymentRequest.ID, err.Error())

						errSaveFailedPaymentRequest := s.Database.UpdateFailedPaymentRequest(ctx, paymentRequest.ID, err.Error())
						if errSaveFailedPaymentRequest != nil {
							log.Printf("[req id: %d] s.Database.UpdateFailedPaymentRequest() got error: %v", paymentRequest.ID, errSaveFailedPaymentRequest)
						}

						continue
					}

					// update status payment request to SUCCESS
					err = s.Database.UpdateSuccessPaymentRequest(ctx, paymentRequest.ID)
					if err != nil {
						log.Printf("[req id: %d] s.Database.UpdateSuccessPaymentRequest() got error: %v", paymentRequest.ID, err)
						continue
					}

					// save data to table payment
					err = s.Database.SavePayment(ctx, models.Payment{
						OrderID:     paymentRequest.OrderID,
						UserID:      paymentRequest.UserID,
						Amount:      paymentRequest.Amount,
						ExternalID:  externalID,
						Status:      "PENDING",
						ExpiredTime: xenditInvoiceRes.ExpiryDate,
						CreateTime:  time.Now(),
					})
					if err != nil {
						log.Printf("[req id: %d] s.Database.SavePayment() got error: %v", paymentRequest.ID, err)
						continue
					}
				}
			}

			time.Sleep(5 * time.Second) // give time gap before next iteration
		}
	}(ctx)
}

func (s *SchedulerService) StartProcessFailedPaymentRequests() {
	go func(ctx context.Context) {
		for {
			// get list of failed payment requests
			var paymentRequests []models.PaymentRequests
			err := s.Database.GetFailedPaymentRequests(ctx, &paymentRequests)
			if err != nil {
				log.Printf("s.Database.GetFailedPaymentRequests() got error: %v", err)
				time.Sleep(10 * time.Second) // give time gap before next iteration
				continue
			}

			// update status to PENDING
			for _, paymentRequest := range paymentRequests {
				err = s.Database.UpdatePendingPaymentRequest(ctx, paymentRequest.ID)
				if err != nil {
					log.Printf("s.Database.UpdatePendingPaymentRequest() got error: %v", err)

					// another retry process
					errUpdateStatus := s.Database.UpdateFailedPaymentRequest(ctx, paymentRequest.ID, err.Error())
					if errUpdateStatus != nil {
						log.Printf("s.Database.UpdateFailedPaymentRequest() got error: %v", errUpdateStatus.Error())
					}

					continue
				}
			}

			time.Sleep(1 * time.Minute) // give time gap before next iteration
		}
	}(context.Background())
}

func (s *SchedulerService) StartCheckPendingInvoices() {
	ticker := time.NewTicker(10 * time.Minute)

	go func() {
		for range ticker.C {
			// query pending invoices
			ctx := context.Background()
			listPendingInvoices, err := s.Database.GetPendingInvoices(ctx)
			if err != nil {
				log.Printf("s.Database.GetPendingInvoices() got error: %v", err)
				continue
			}

			for _, pendingInvoice := range listPendingInvoices {
				invoiceStatus, err := s.Xendit.CheckInvoiceStatus(ctx, pendingInvoice.ExternalID)
				if err != nil {
					log.Printf("s.Xendit.CheckInvoiceStatus() got error: %v", err)
					continue
				}

				if invoiceStatus == "PAID" {
					err = s.PaymentService.ProcessPaymentSuccess(ctx, pendingInvoice.OrderID)
					if err != nil {
						log.Printf("s.PaymentService.ProcessPaymentSuccess() got error: %v", err)
						continue
					}
				}
			}
		}
	}()
}
