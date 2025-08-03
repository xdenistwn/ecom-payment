package main

import (
	"context"
	"payment/cmd/payment/handler"
	"payment/cmd/payment/repository"
	"payment/cmd/payment/resource"
	"payment/cmd/payment/service"
	"payment/cmd/payment/usecase"
	"payment/config"
	"payment/infrastructure/constant"
	"payment/infrastructure/log"
	"payment/kafka"
	"payment/models"
	"payment/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// init config
	cfg := config.LoadConfig()
	// init connection
	db := resource.InitDb(&cfg)
	kafkaWriter := kafka.NewWriter(cfg.Kafka.Broker, cfg.Kafka.Topics[constant.KafkaTopicPaymentSuccess])

	// setup logger
	log.SetupLogger()

	// payment service
	databaseRepository := repository.NewPaymentDatabase(db)
	publisherRepository := repository.NewKafkaPublisher(kafkaWriter)
	paymentService := service.NewPaymentService(databaseRepository, publisherRepository)
	paymentUsecase := usecase.NewPaymentUsecase(paymentService)
	paymentHandler := handler.NewPaymentHandler(paymentUsecase, cfg.Xendit.WebhookToken)

	// xendit service
	xenditRepository := repository.NewXenditClient(cfg.Xendit.SecretApiKey)
	xenditService := service.NewXenditService(databaseRepository, xenditRepository)
	xenditUsacase := usecase.NewXenditUsecase(xenditService)

	// scheduler service
	schedulerService := service.SchedulerService{
		Database:       databaseRepository,
		Xendit:         xenditRepository,
		Publisher:      publisherRepository,
		PaymentService: paymentService,
	}

	// start scheduler
	schedulerService.StartCheckPendingInvoices()
	schedulerService.StartProcessPendingPaymentRequests()
	schedulerService.StartProcessFailedPaymentRequests()
	schedulerService.StartProcessExpiredPendingPayments()

	// kafka consumer
	// potential not effienct when traffic is high, consider using a more robust solution like a message queue
	kafka.StartOrderConsumer(cfg.Kafka.Broker, cfg.Kafka.Topics[constant.KafkaTopicOrderCreated],
		func(event models.OrderCreatedEvent) {
			// async process
			if cfg.Toggle.DisableCreateInvoiceDirectly {
				if err := paymentUsecase.ProcessPaymentRequest(context.Background(), event); err != nil {
					log.Logger.Println("Failed handling order_created event:", err.Error())
				}
			} else { // sync process
				if err := xenditUsacase.CreateInvoice(context.Background(), event); err != nil {
					log.Logger.Println("Failed handling order_created event", err.Error())
				}
			}
		})

	port := cfg.App.Port
	router := gin.Default()
	routes.SetupRoutes(router, paymentHandler)

	router.Run(":" + port)

	log.Logger.Printf("Server listening on port: %s", port)
}
