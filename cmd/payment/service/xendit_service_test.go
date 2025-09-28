package service

import (
	"context"
	"fmt"
	mocks "payment/cmd/test_mock"
	"payment/infrastructure/log"
	"payment/models"
	"payment/proto/userpb"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_CreateInvoice(t *testing.T) {
	type mockFields struct {
		userClient *mocks.MockUserClient
		xendit     *mocks.MockXenditClient
		database   *mocks.MockPaymentDatabase
	}

	type args struct {
		ctx   context.Context
		param models.OrderCreatedEvent
	}

	mockTime := time.Now()

	tests := []struct {
		name      string
		args      args
		mock      func(mockFields)
		wantError error
	}{
		{
			name: "given_valid_param_but_got_error_GetUserInfoByUserID_then_it_should_return_error_s.CI001",
			args: args{
				ctx: context.Background(),
				param: models.OrderCreatedEvent{
					OrderID:         1,
					UserID:          12345,
					TotalAmount:     10000,
					PaymentMethod:   "GoPay",
					ShippingAddress: "Jl. Elang Testing 123",
				},
			},
			mock: func(mf mockFields) {
				mf.userClient.EXPECT().GetUserInfoByUserId(context.Background(), int64(12345)).Return(&userpb.GetUserInfoResult{}, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "given_valid_param_and_success_GetUserInfoByUserID_but_got_error_CreateInvoice_then_it_should_return_error_s.CI002",
			args: args{
				ctx: context.Background(),
				param: models.OrderCreatedEvent{
					OrderID:         123,
					UserID:          12345,
					TotalAmount:     10000,
					PaymentMethod:   "GoPay",
					ShippingAddress: "Jl. Elang Testing 123",
				},
			},
			mock: func(mf mockFields) {
				mf.userClient.EXPECT().GetUserInfoByUserId(context.Background(), int64(12345)).Return(&userpb.GetUserInfoResult{
					Id:    12345,
					Name:  "Deni Setiawan",
					Email: "ofc.denisetiawan@gmail.com",
					Role:  "admin",
				}, nil)

				mf.xendit.EXPECT().CreateInvoice(context.Background(), models.XenditInvoiceRequest{
					ExternalID:  fmt.Sprintf("order-%d", 123),
					Amount:      10000,
					Description: fmt.Sprintf("Pembayaran Order %d", 123),
					PayerEmail:  "ofc.denisetiawan@gmail.com",
				}).Return(models.XenditInvoiceResponse{}, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "given_valid_param_and_success_GetUserInfoByUserID_then_success_CreateInvoice_but_got_error_SavePayment_then_it_should_return_error_s.CI003",
			args: args{
				ctx: context.Background(),
				param: models.OrderCreatedEvent{
					OrderID:         111,
					UserID:          222,
					TotalAmount:     3000,
					PaymentMethod:   "OVO",
					ShippingAddress: "Jl. Elang Testing 123",
				},
			},
			mock: func(mf mockFields) {
				mf.userClient.EXPECT().GetUserInfoByUserId(context.Background(), int64(222)).Return(&userpb.GetUserInfoResult{
					Id:    222,
					Name:  "Deni Setiawan",
					Email: "ofc.denisetiawan@gmail.com",
					Role:  "user",
				}, nil)

				mf.xendit.EXPECT().CreateInvoice(context.Background(), models.XenditInvoiceRequest{
					ExternalID:  fmt.Sprintf("order-%d", 111),
					Amount:      3000,
					Description: fmt.Sprintf("Pembayaran Order %d", 111),
					PayerEmail:  "ofc.denisetiawan@gmail.com",
				}).Return(models.XenditInvoiceResponse{
					ID:         "xendit-invoice_111",
					ExpiryDate: mockTime.AddDate(0, 0, 3),
					InvoiceURL: "/payment/invoice?id=xendit-invoice_111",
					Status:     "PENDING",
				}, nil)

				mf.database.EXPECT().SavePayment(context.Background(), gomock.Any()).Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "given_valid_param_and_success_GetUserInfoByUserID_then_success_CreateInvoice_but_got_error_SavePayment_then_it_should_return_nil_error",
			args: args{
				ctx: context.Background(),
				param: models.OrderCreatedEvent{
					OrderID:         111,
					UserID:          222,
					TotalAmount:     3000,
					PaymentMethod:   "OVO",
					ShippingAddress: "Jl. Elang Testing 123",
				},
			},
			mock: func(mf mockFields) {
				mf.userClient.EXPECT().GetUserInfoByUserId(context.Background(), int64(222)).Return(&userpb.GetUserInfoResult{
					Id:    222,
					Name:  "Deni Setiawan",
					Email: "ofc.denisetiawan@gmail.com",
					Role:  "user",
				}, nil)

				mf.xendit.EXPECT().CreateInvoice(context.Background(), models.XenditInvoiceRequest{
					ExternalID:  fmt.Sprintf("order-%d", 111),
					Amount:      3000,
					Description: fmt.Sprintf("Pembayaran Order %d", 111),
					PayerEmail:  "ofc.denisetiawan@gmail.com",
				}).Return(models.XenditInvoiceResponse{
					ID:         "xendit-invoice_111",
					ExpiryDate: mockTime.AddDate(0, 0, 3),
					InvoiceURL: "/payment/invoice?id=xendit-invoice_111",
					Status:     "PENDING",
				}, nil)

				mf.database.EXPECT().SavePayment(context.Background(), gomock.Any()).Return(nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			log.SetupLogger()

			mock := mockFields{
				userClient: mocks.NewMockUserClient(ctrl),
				xendit:     mocks.NewMockXenditClient(ctrl),
				database:   mocks.NewMockPaymentDatabase(ctrl),
			}

			service := &xenditService{
				userClient: mock.userClient,
				database:   mock.database,
				xendit:     mock.xendit,
			}

			test.mock(mock)
			gotError := service.CreateInvoice(test.args.ctx, test.args.param)
			assert.Equal(t, gotError, test.wantError)
		})
	}
}
