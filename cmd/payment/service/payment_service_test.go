package service

import (
	"context"
	mocks "payment/cmd/test_mock"
	mocksRepository "payment/cmd/test_mock"
	"payment/infrastructure/log"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_CheckPaymentAmountByOrderID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// mockPaymentService := mocksService.NewMockPaymentService(ctrl)
	mockRepositoryDatabase := mocksRepository.NewMockPaymentDatabase(ctrl)
	mockRepositoryPublisher := mocksRepository.NewMockPaymentEventPublisher(ctrl)

	// expected result
	var expecetedAmount float64 = 10000.0

	// mock the Repository because payment service depends on it
	mockRepositoryDatabase.EXPECT().CheckPaymentAmountByOrderID(context.Background(), int64(1)).Return(expecetedAmount, nil)

	// actual result
	paymentService := paymentService{
		database:  mockRepositoryDatabase,
		publisher: mockRepositoryPublisher,
	}
	actualAmount, err := paymentService.CheckPaymentAmountByOrderID(context.Background(), int64(1))
	assert.Equal(t, actualAmount, expecetedAmount)
	assert.NoError(t, err)
}

func Test_CheckPaymentAmountByOrderID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// log logger
	log.SetupLogger()

	mockRepositoryDatabase := mocksRepository.NewMockPaymentDatabase(ctrl)
	mockRepositoryPublisher := mocksRepository.NewMockPaymentEventPublisher(ctrl)

	// expected result
	var expecetedAmount float64 = 0.0

	// mock the Repository because payment service depends on it
	mockRepositoryDatabase.EXPECT().CheckPaymentAmountByOrderID(context.Background(), int64(1)).Return(expecetedAmount, assert.AnError)

	// actual result
	paymentService := paymentService{
		database:  mockRepositoryDatabase,
		publisher: mockRepositoryPublisher,
	}
	actualAmount, err := paymentService.CheckPaymentAmountByOrderID(context.Background(), int64(1))
	assert.Equal(t, actualAmount, expecetedAmount)
	assert.Error(t, err)
}

func Test_CheckPaymentAmountByOrderID(t *testing.T) {
	type mockFields struct {
		database  *mocks.MockPaymentDatabase
		publisher *mocks.MockPaymentEventPublisher
	}

	type args struct {
		ctx     context.Context
		orderID int64
	}

	// log logger
	log.SetupLogger()

	// expeceted result
	expectedAmountSuccess := float64(10000)
	expectedAmountError := float64(0)

	// list of test cases
	tests := []struct {
		name       string
		args       args
		mock       func(fields mockFields)
		wantResult float64
		wantError  error
	}{
		{
			name: "Success",
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			mock: func(fields mockFields) {
				fields.database.EXPECT().CheckPaymentAmountByOrderID(context.Background(), int64(1)).Return(expectedAmountSuccess, nil)
			},
			wantResult: expectedAmountSuccess,
			wantError:  nil,
		},
		{
			name: "Error",
			args: args{
				ctx:     context.Background(),
				orderID: 1,
			},
			mock: func(fields mockFields) {
				fields.database.EXPECT().CheckPaymentAmountByOrderID(context.Background(), int64(1)).Return(expectedAmountError, assert.AnError)
			},
			wantResult: expectedAmountError,
			wantError:  assert.AnError,
		},
	}

	// loop through each test case
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create mock objects
			mockFields := mockFields{
				database:  mocks.NewMockPaymentDatabase(ctrl),
				publisher: mocks.NewMockPaymentEventPublisher(ctrl),
			}

			// apply the mock expectations
			test.mock(mockFields)

			// create the payment service with mocked dependencies
			paymentService := paymentService{
				database:  mockFields.database,
				publisher: mockFields.publisher,
			}

			// call the method under test
			result, err := paymentService.CheckPaymentAmountByOrderID(test.args.ctx, test.args.orderID)

			// assert the results
			assert.Equal(t, result, test.wantResult)
			assert.Equal(t, err, test.wantError)
		})
	}
}
