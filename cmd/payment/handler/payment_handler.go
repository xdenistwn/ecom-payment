package handler

import (
	"net/http"
	"payment/cmd/payment/usecase"
	"payment/infrastructure/log"
	"payment/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PaymentHandler interface {
	HandleXenditWebhook(c *gin.Context)
	HandlerDownloadPDFInvoice(c *gin.Context)
}

type paymentHandler struct {
	Usecase            usecase.PaymentUsecase
	XenditWebhookToken string
}

func NewPaymentHandler(usecase usecase.PaymentUsecase, xenditWebhookToken string) PaymentHandler {
	return &paymentHandler{
		Usecase:            usecase,
		XenditWebhookToken: xenditWebhookToken,
	}
}

func (h *paymentHandler) HandleXenditWebhook(c *gin.Context) {
	var payload models.XenditWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Logger.WithFields(logrus.Fields{
			"payload": payload,
		})

		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "Invalid payload",
			"error_message": err.Error(),
		})

		return
	}

	// validate webhook token
	headerWebhookToken := c.GetHeader("x-callback-token")
	if h.XenditWebhookToken != headerWebhookToken {
		log.Logger.WithFields(logrus.Fields{
			"xendit_callback_webhook_token": headerWebhookToken,
		}).Errorf("Invalid Webhook token: %s", headerWebhookToken)

		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid webhook token",
		})

		return
	}

	err := h.Usecase.ProcessPaymentWebhook(c.Request.Context(), payload)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"payload": payload,
		})

		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success.",
	})

	return
}

func (h *paymentHandler) HandlerCreateInvoice(c *gin.Context) {
	var payload models.OrderCreatedEvent
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})

		return
	}

	return
}

func (h *paymentHandler) HandlerDownloadPDFInvoice(c *gin.Context) {
	orderIDStr := c.Param("order_id")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Order ID is required",
		})

		return
	}

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)

	filePath, err := h.Usecase.DownloadPDFInvoice(c.Request.Context(), orderID)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"order_id": orderID,
		}).Errorf("DownloadPDFInvoice got error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to download PDF invoice",
		})

		return
	}

	c.FileAttachment(filePath, filePath)
}
