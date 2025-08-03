package routes

import (
	"payment/cmd/payment/handler"
	"payment/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, paymentHandler handler.PaymentHandler) {
	// context timeout and logger
	router.Use(middleware.RequestLogger(2))
	router.POST("/v1/payment/webhook", paymentHandler.HandleXenditWebhook)
	router.GET("/v1/payment/invoice/:order_id/pdf", paymentHandler.HandlerDownloadPDFInvoice)
}
